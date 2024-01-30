package controller

import (
	"context"
	"log"
	"time"

	snap "github.com/kubernetes-csi/external-snapshotter/client/v6/apis/volumesnapshot/v1"
	exss "github.com/kubernetes-csi/external-snapshotter/client/v6/clientset/versioned"
	klientSet "github.com/saima-s/data-restore/pkg/client/clientset/versioned"
	kinf "github.com/saima-s/data-restore/pkg/client/informers/externalversions/saima.dev.com/v1"
	"golang.org/x/time/rate"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
)

type SnapshotController struct {
	klient            klientSet.Interface
	informer          cache.SharedIndexInformer
	queue             workqueue.RateLimitingInterface
	snapshotterClient exss.Interface
}

func NewSnapshotController(klient klientSet.Interface, snapshotterClient exss.Interface, informer kinf.DataSnapshotInformer) *SnapshotController {

	inf := informer.Informer()
	ratelimiter := workqueue.NewMaxOfRateLimiter(
		workqueue.NewItemExponentialFailureRateLimiter(5*time.Millisecond, 1000*time.Second),
		&workqueue.BucketRateLimiter{Limiter: rate.NewLimiter(rate.Limit(50), 300)},
	)

	ctrl := &SnapshotController{
		klient:            klient,
		informer:          inf,
		queue:             workqueue.NewRateLimitingQueue(ratelimiter),
		snapshotterClient: snapshotterClient,
	}

	inf.AddEventHandler(
		cache.ResourceEventHandlerFuncs{
			AddFunc: ctrl.enqueueTask,

			UpdateFunc: func(old, new interface{}) {
				ctrl.enqueueTask(new)
			},
		},
	)
	return ctrl

}

func (c *SnapshotController) enqueueTask(obj interface{}) {

	log.Println("EnqueueTask was called")

	c.queue.Add(obj)
}

func (c *SnapshotController) Run(ch <-chan struct{}) {
	log.Println("Starting snapshot controller")
	if !cache.WaitForCacheSync(ch, c.informer.HasSynced) {
		log.Println("Waiting for cache to be synced")
	}

	go wait.Until(c.snapshotWorker, 1*time.Second, ch)

	<-ch
}

func (c *SnapshotController) snapshotWorker() {
	for c.processItem() {

	}
}

func (c *SnapshotController) processItem() bool {
	obj, shutdown := c.queue.Get()
	if shutdown {
		return false
	}
	defer func() {
		c.queue.Forget(obj)
		c.queue.ShutDown()
	}()

	key, err := cache.MetaNamespaceKeyFunc(obj)
	if err != nil {
		log.Printf("Error getting key from cache %v", err.Error())
	}

	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		log.Printf("Error splitting key into namespace and name: %v", err.Error())
		return false
	}

	err = c.Backup(namespace, name)
	if err != nil {
		return false
	}

	defer c.queue.Done(obj)

	return true
}

func (c *SnapshotController) Backup(namespace, name string) error {
	resource, err := c.klient.SaimaV1().DataSnapshots(namespace).Get(context.Background(), name, metav1.GetOptions{})
	if err != nil {
		log.Printf("Error in getting the snapshot custom resouce: %v", err.Error())
		return err
	}

	volumeSnapshotClass := resource.Spec.VolumeSnapshotClass
	snapshot := snap.VolumeSnapshot{
		ObjectMeta: metav1.ObjectMeta{
			Name:      resource.Spec.VolumeSnapshot,
			Namespace: resource.Spec.Namespace,
		},
		Spec: snap.VolumeSnapshotSpec{
			VolumeSnapshotClassName: &volumeSnapshotClass,
			Source: snap.VolumeSnapshotSource{
				PersistentVolumeClaimName: &resource.Spec.PvcName,
			},
		},
	}

	snap, err := c.snapshotterClient.SnapshotV1().VolumeSnapshots(resource.Spec.Namespace).Get(context.Background(), resource.Spec.VolumeSnapshot, metav1.GetOptions{})
	if err != nil {
		log.Printf("Error in getting snapshot for %v", err.Error())
	}

	if snap != nil {
		log.Printf("Snapshot already exists with this name: %v", snap.Name)

		return nil
	}

	_, err = c.snapshotterClient.SnapshotV1().VolumeSnapshots(resource.Spec.Namespace).Create(context.Background(), &snapshot, metav1.CreateOptions{})
	if err != nil {
		log.Printf("Error in creating snapshot for %v", err.Error())
		return err
	}
	log.Printf("Snapshot Created for %s", name)
	return nil
}
