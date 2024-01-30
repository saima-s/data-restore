/*
Copyright The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Code generated by lister-gen. DO NOT EDIT.

package v1

import (
	v1 "github.com/saima-s/data-restore/pkg/apis/saima.dev.com/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/cache"
)

// DataRestoreLister helps list DataRestores.
// All objects returned here must be treated as read-only.
type DataRestoreLister interface {
	// List lists all DataRestores in the indexer.
	// Objects returned here must be treated as read-only.
	List(selector labels.Selector) (ret []*v1.DataRestore, err error)
	// DataRestores returns an object that can list and get DataRestores.
	DataRestores(namespace string) DataRestoreNamespaceLister
	DataRestoreListerExpansion
}

// dataRestoreLister implements the DataRestoreLister interface.
type dataRestoreLister struct {
	indexer cache.Indexer
}

// NewDataRestoreLister returns a new DataRestoreLister.
func NewDataRestoreLister(indexer cache.Indexer) DataRestoreLister {
	return &dataRestoreLister{indexer: indexer}
}

// List lists all DataRestores in the indexer.
func (s *dataRestoreLister) List(selector labels.Selector) (ret []*v1.DataRestore, err error) {
	err = cache.ListAll(s.indexer, selector, func(m interface{}) {
		ret = append(ret, m.(*v1.DataRestore))
	})
	return ret, err
}

// DataRestores returns an object that can list and get DataRestores.
func (s *dataRestoreLister) DataRestores(namespace string) DataRestoreNamespaceLister {
	return dataRestoreNamespaceLister{indexer: s.indexer, namespace: namespace}
}

// DataRestoreNamespaceLister helps list and get DataRestores.
// All objects returned here must be treated as read-only.
type DataRestoreNamespaceLister interface {
	// List lists all DataRestores in the indexer for a given namespace.
	// Objects returned here must be treated as read-only.
	List(selector labels.Selector) (ret []*v1.DataRestore, err error)
	// Get retrieves the DataRestore from the indexer for a given namespace and name.
	// Objects returned here must be treated as read-only.
	Get(name string) (*v1.DataRestore, error)
	DataRestoreNamespaceListerExpansion
}

// dataRestoreNamespaceLister implements the DataRestoreNamespaceLister
// interface.
type dataRestoreNamespaceLister struct {
	indexer   cache.Indexer
	namespace string
}

// List lists all DataRestores in the indexer for a given namespace.
func (s dataRestoreNamespaceLister) List(selector labels.Selector) (ret []*v1.DataRestore, err error) {
	err = cache.ListAllByNamespace(s.indexer, s.namespace, selector, func(m interface{}) {
		ret = append(ret, m.(*v1.DataRestore))
	})
	return ret, err
}

// Get retrieves the DataRestore from the indexer for a given namespace and name.
func (s dataRestoreNamespaceLister) Get(name string) (*v1.DataRestore, error) {
	obj, exists, err := s.indexer.GetByKey(s.namespace + "/" + name)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.NewNotFound(v1.Resource("datarestore"), name)
	}
	return obj.(*v1.DataRestore), nil
}