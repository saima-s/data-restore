package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/client-go/kubernetes"

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

	kubeConfig := flag.String("kubeconfig", "/home/saima/.kube/config", "location to kube config file")
	config, err := clientcmd.BuildConfigFromFlags("", *kubeConfig)
	if err != nil {
		fmt.Println("error in building config is:", err.Error())
		config, err = rest.InClusterConfig()
		if err != nil {
			fmt.Println("error in gettin config from cluster is:", err.Error())
		}
	}

	config, err = clientcmd.BuildConfigFromFlags("", *kubeConfig)
	if err != nil {
		log.Printf("Building config from flags failed, %s, trying to build inclusterconfig", err.Error())
		config, err = rest.InClusterConfig()
		if err != nil {
			log.Printf("error %s building inclusterconfig", err.Error())
		}
	}

	klientSet, err := klient.NewForConfig(config)
	if err != nil {
		log.Printf("Error %s getting clientset", err.Error())
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

	_, err = klientSet.SaimaV1().DataRestores("default").List(context.Background(), metav1.ListOptions{})
	if err != nil {
		log.Printf("Error %s getting list of dataRestorePvc", err.Error())
	}

	infoFactory := kInfFac.NewSharedInformerFactory(klientSet, 20*time.Minute)
	snapshotctrl := controller.NewSnapshotController(klientSet, snapClient, infoFactory.Saima().V1().DataSnapshots())
	restorectrl := controller.NewRestoreController(klientSet, restoreClient, infoFactory.Saima().V1().DataRestores())
	ch := make(<-chan struct{})
	go infoFactory.Start(ch)
	go snapshotctrl.Run(ch)
	restorectrl.Run(ch)

}
