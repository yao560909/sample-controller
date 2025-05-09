package main

import (
	"flag"
	clientset "github.com/yao560909/sample-controller/pkg/generated/clientset/versioned"
	informers "github.com/yao560909/sample-controller/pkg/generated/informers/externalversions"
	"github.com/yao560909/sample-controller/pkg/signals"
	kubeinformers "k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"
	"time"
)

var (
	masterURL  string
	kubeconfig string
)

func init() {
	flag.StringVar(&kubeconfig, "kubeconfig", "", "Path to a kubeconfig. Only required if out-of-cluster.")
	flag.StringVar(&masterURL, "master", "", "The address of the Kubernetes API server. Overrides any value in kubeconfig. Only required if out-of-cluster.")
}

func main() {
	klog.InitFlags(nil)
	flag.Parse()
	ctx := signals.SetupSignalHandler()
	logger := klog.FromContext(ctx)
	restConfig, err := clientcmd.BuildConfigFromFlags(masterURL, kubeconfig)
	if err != nil {
		logger.Error(err, "Error building kubeconfig")
		klog.FlushAndExit(klog.ExitFlushTimeout, 1)
	}
	kubeClient, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		logger.Error(err, "Error building kubernetes clientset")
		klog.FlushAndExit(klog.ExitFlushTimeout, 1)
	}
	exampleClient, err := clientset.NewForConfig(restConfig)
	if err != nil {
		logger.Error(err, "Failed to build Kubernetes clientset for CRDs")
		klog.FlushAndExit(klog.ExitFlushTimeout, 1)
	}
	kubeSharedInformerFactory := kubeinformers.NewSharedInformerFactory(kubeClient, time.Second*30)
	exampleSharedInformerFactory := informers.NewSharedInformerFactory(exampleClient, time.Second*30)

	controller := NewController(ctx, kubeClient, exampleClient,
		kubeSharedInformerFactory.Apps().V1().Deployments(),
		exampleSharedInformerFactory.Samplecontroller().V1alpha1().Foos())
	// notice that there is no need to run Start methods in a separate goroutine. (i.e. go kubeInformerFactory.Start(ctx.done())
	// Start method is non-blocking and runs all registered informers in a dedicated goroutine.
	kubeSharedInformerFactory.Start(ctx.Done())
	exampleSharedInformerFactory.Start(ctx.Done())

	if err = controller.Run(ctx, 2); err != nil {
		logger.Error(err, "Error running controller")
		klog.FlushAndExit(klog.ExitFlushTimeout, 1)
	}

}
