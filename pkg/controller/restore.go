package controller

import (
	"context"
	"log"
	"time"

	klientSet "github.com/saima-s/data-restore/pkg/client/clientset/versioned"
	kinf "github.com/saima-s/data-restore/pkg/client/informers/externalversions/saima.dev.com/v1"
	"golang.org/x/time/rate"
	corev1 "k8s.io/api/core/v1"
	parser "k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
)

var (
	apiGroup           = "snapshot.storage.k8s.io"
	volumeSnapshotKind = "VolumeSnapshot"
)

type RestoreController struct {
	klient        klientSet.Interface
	informer      cache.SharedIndexInformer
	queue         workqueue.RateLimitingInterface
	restoreClient kubernetes.Interface
}

func NewRestoreController(klient klientSet.Interface, restoreClient kubernetes.Interface, informer kinf.DataRestoreInformer) *RestoreController {

	inf := informer.Informer()
	ratelimiter := workqueue.NewMaxOfRateLimiter(
		workqueue.NewItemExponentialFailureRateLimiter(5*time.Millisecond, 1000*time.Second),
		&workqueue.BucketRateLimiter{Limiter: rate.NewLimiter(rate.Limit(50), 300)},
	)

	ctrl := &RestoreController{
		klient:        klient,
		informer:      inf,
		queue:         workqueue.NewRateLimitingQueue(ratelimiter),
		restoreClient: restoreClient,
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

func (c *RestoreController) enqueueTask(obj interface{}) {
	log.Println("restore enqueueTask was called")

	c.queue.Add(obj)
}

func (c *RestoreController) Run(ch <-chan struct{}) {

	log.Println("starting restore controller")
	if !cache.WaitForCacheSync(ch, c.informer.HasSynced) {
		log.Print("waiting for cache to be synced\n")
	}

	go wait.Until(c.restoreWorker, 1*time.Second, ch)

	<-ch
}

func (c *RestoreController) restoreWorker() {
	for c.processRestoreItem() {

	}
}

func (c *RestoreController) processRestoreItem() bool {
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
		log.Printf("Eerror getting key from cache %v", err.Error())
	}

	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		log.Printf("error splitting key into namespace and name: %v", err.Error())
		return false
	}

	err = c.Restore(namespace, name)
	if err != nil {
		return false
	}

	defer c.queue.Done(obj)

	return true
}

func (c *RestoreController) Restore(namespace, name string) error {
	resource, err := c.klient.SaimaV1().DataRestores(namespace).Get(context.Background(), name, metav1.GetOptions{})
	if err != nil {
		log.Printf("error in getting the custom resouce: %v", err.Error())
		return err
	}

	volumeSnapshotClass := resource.Spec.VolumeSnapshotClass
	pvc := corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name: resource.Name,
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			StorageClassName: &volumeSnapshotClass,
			DataSource: &corev1.TypedLocalObjectReference{
				Name:     resource.Spec.SnapshotName,
				Kind:     volumeSnapshotKind,
				APIGroup: &apiGroup,
			},
			AccessModes: []corev1.PersistentVolumeAccessMode{
				corev1.ReadWriteOnce,
			},
			Resources: corev1.VolumeResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceStorage: parser.MustParse(resource.Spec.Storage),
				},
			},
		},
	}

	p, err := c.restoreClient.CoreV1().PersistentVolumeClaims(resource.Namespace).Get(context.Background(), resource.Name, metav1.GetOptions{})
	if err != nil {
		log.Printf("error in getting PVC is: %v", err.Error())
	}

	if p != nil {
		log.Printf("this PVC already exists:%v", p.Name)
		return nil
	}

	pvcCreated, err := c.restoreClient.CoreV1().PersistentVolumeClaims(resource.Namespace).Create(context.Background(), &pvc, metav1.CreateOptions{})
	if err != nil {
		log.Printf("error in restoring data is: %v", err.Error())
		return err
	}

	log.Printf("Data Restored %s", pvcCreated)
	return nil
}
