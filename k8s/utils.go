// Copyright 2017 AMIS Technologies
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package k8s

import (
	"os"
	"path/filepath"
	"sync"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

const (
	defaultNamespace = "default"

	healthCheckRetryCount = 5
	healthCheckRetryDelay = 5 * time.Second
)

func k8sClient(podName string) *kubernetes.Clientset {
	config, err := clientcmd.BuildConfigFromFlags("", filepath.Join(os.Getenv("HOME"), ".kube", "config"))
	if err != nil {
		log.Error("Failed to create Kubernetes config", "err", err)
		return nil
	}

	for i := 0; i < healthCheckRetryCount; i++ {
		client, err := kubernetes.NewForConfig(config)
		if err != nil {
			log.Error("Failed to create Kubernetes client from config", "config", config, "err", err)
			<-time.After(healthCheckRetryDelay)
			continue
		}
		_, err = client.CoreV1().Pods(defaultNamespace).Get(podName, metav1.GetOptions{})
		if err != nil {
			log.Error("Failed to get pod", "namespace", defaultNamespace, "pod", podName, "err", err)
			<-time.After(healthCheckRetryDelay)
			continue
		} else {
			return client
		}
	}

	log.Error("Failed to retrieve kubernetes client")
	return nil
}

func executeInParallel(fns ...func() error) error {
	var wg sync.WaitGroup
	errc := make(chan error, len(fns))
	wg.Add(len(fns))

	for _, fn := range fns {
		fn := fn
		go func() {
			defer wg.Done()
			errc <- fn()
		}()
	}
	// Wait for the first error, then terminate the others.
	var err error
	for i := 0; i < len(fns); i++ {
		if err = <-errc; err != nil {
			break
		}
	}
	wg.Wait()
	return err
}
