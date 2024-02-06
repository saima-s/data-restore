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
		},
	)
	return ctrl

}

func (c *SnapshotController) enqueueTask(obj interface{}) {

	log.Printf("EnqueueTask was called: %v", obj)

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

	namespace, crName, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		log.Printf("Error splitting key into namespace and name: %v", err.Error())
		return false
	}

	err = c.Backup(namespace, crName)
	if err != nil {
		return false
	}

	defer c.queue.Done(obj)

	return true
}

func (c *SnapshotController) Backup(namespace, crName string) error {
	resource, err := c.klient.SaimaV1().DataSnapshots(namespace).Get(context.Background(), crName, metav1.GetOptions{})
	if err != nil {
		log.Printf("Error in getting the snapshot custom resouce: %v", err.Error())
		return err
	}

	volumeSnapshotClass := resource.Spec.VolumeSnapshotClass
	snapshot := snap.VolumeSnapshot{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: "snapshot-",
			Namespace:    namespace, // create a snapshot in namespace where CR is present(here datasnapshot-0)
		},
		Spec: snap.VolumeSnapshotSpec{
			VolumeSnapshotClassName: &volumeSnapshotClass,
			Source: snap.VolumeSnapshotSource{
				PersistentVolumeClaimName: &resource.Spec.PvcName,
			},
		},
	}

	snap, err := c.snapshotterClient.SnapshotV1().VolumeSnapshots(namespace).Create(context.Background(), &snapshot, metav1.CreateOptions{})
	if err != nil {
		log.Printf("Error in creating snapshot for %v", err.Error())
		return err
	}

	generatedName := snap.Name
	err = c.updateStatus(creating, generatedName, crName, namespace)
	if err != nil {
		log.Printf("Error in updating snapshot %s", err)
		return err
	}

	go c.waitForBackup(generatedName, namespace, crName, namespace)
	log.Printf("Snapshot Created for %s", crName)
	return nil
}

func (c *SnapshotController) updateStatus(progress, snapshotName, crName, namespace string) error {
	resource, err := c.klient.SaimaV1().DataSnapshots(namespace).Get(context.Background(), crName, metav1.GetOptions{})
	if err != nil {
		log.Printf("Error in getting the snapshot custom resouce for updating status: %v", err.Error())
		return err
	}

	newResource := resource.DeepCopy()
	newResource.Status.Progress = progress
	newResource.Status.SnapshotName = snapshotName

	_, err = c.klient.SaimaV1().DataSnapshots(namespace).UpdateStatus(context.Background(), newResource, metav1.UpdateOptions{})
	if err != nil {
		log.Printf("Error in updating the status of snapshot: %v", err)
		return err
	}

	return nil

}

func (c *SnapshotController) waitForBackup(snapshotName string, snapNamespace string, crName string, crNamespace string) {
	err := wait.Poll(5*time.Second, 5*time.Minute, func() (done bool, err error) {
		status := c.getSnapshotState(snapshotName, snapNamespace)
		if status == true {
			err = c.updateStatus(created, snapshotName, crName, crNamespace)
			if err != nil {
				log.Printf("Error in updating status %s", err.Error())
			}

			log.Printf("Updating status to created")
			return true, nil
		}

		log.Println("Waiting for volume snapshot to get ready")
		return false, nil
	})
	if err != nil {
		log.Printf("Error waiting backup for created %s", err.Error())
		return
	}
}

func (c *SnapshotController) getSnapshotState(name, namespace string) bool {
	backup, err := c.snapshotterClient.SnapshotV1().VolumeSnapshots(namespace).Get(context.Background(), name, metav1.GetOptions{})
	if err != nil {
		log.Printf("Error Getting Current State of Volume Snapshot %s.\nReason --> %s", name, err.Error())
	}

	status := backup.Status.ReadyToUse
	return *status
}
