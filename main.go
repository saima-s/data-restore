package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"path/filepath"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/util/homedir"

	exss "github.com/kubernetes-csi/external-snapshotter/client/v6/clientset/versioned"
	klient "github.com/saima-s/data-restore/pkg/client/clientset/versioned"
	kInfFac "github.com/saima-s/data-restore/pkg/client/informers/externalversions"
	"github.com/saima-s/data-restore/pkg/controller"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

var (
	resyncPeriod = flag.Duration("resync-period", 15*time.Minute, "Resync interval of the controller.")
)

func main() {

	var kubeconfig *string

	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()

	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		log.Printf("Building config from flags failed, %s, trying to build inclusterconfig", err.Error())
		config, err = rest.InClusterConfig()
		if err != nil {
			log.Printf("error %s building inclusterconfig", err.Error())
		}
	}

	klientSet, err := klient.NewForConfig(config)
	if err != nil {
		log.Printf("error %s getting clientset", err.Error())
	}

	snapClient, err := exss.NewForConfig(config)
	if err != nil {
		log.Printf("Error building snapshot clientset: %s", err.Error())
		return
	}

	restoreClient, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Printf("Error building restore clientset: %s", err.Error())
		return
	}

	// factory := informers.NewSharedInformerFactory(snapClient, *resyncPeriod)

	fmt.Println("klientSet is:", klientSet)

	dataRestorePvc, err := klientSet.SaimaV1().DataRestores("").List(context.Background(), metav1.ListOptions{})
	if err != nil {
		log.Printf("error %s getting list of dataRestorePvc", err.Error())
	}

	fmt.Println(len(dataRestorePvc.Items))

	infoFactory := kInfFac.NewSharedInformerFactory(klientSet, 20*time.Minute)
	snapshotctrl := controller.NewSnapshotController(klientSet, snapClient, infoFactory.Saima().V1().DataRestores())
	restorectrl := controller.NewRestoreController(klientSet, restoreClient, infoFactory.Saima().V1().DataRestores())

	infoFactory.Start(make(<-chan struct{}))
	snapshotctrl.Run(make(<-chan struct{}))
	restorectrl.Run(make(<-chan struct{}))

	// dynClient, err := dynamic.NewForConfig(config)
	// if err != nil {
	// 	fmt.Printf("error %s, getting dyn client\n", err.Error())
	// }

	// _, err = dynClient.Resource(schema.GroupVersionResource{
	// 	Group:    "saima.dev.com",
	// 	Version:  "v1",
	// 	Resource: "dataRestore",
	// }).Namespace("default").Get(context.Background(), "kluster-0", metav1.GetOptions{})
	// if err != nil {
	// 	fmt.Printf("error %s gettign resource from dyn client\n", err.Error())
	// }

}
