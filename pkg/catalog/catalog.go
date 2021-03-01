// main file for catalog package

// The catalog package is responsible for creating the data structure to hold
// the built catalog in memory. It also provides utility functions to load,
// search, and step through the catalog.
//
package catalog

// Package

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"strings"
	"sync"
	"time"

	"example.com/jcfg/pkg/api"
	"example.com/jcfg/pkg/resources"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type Catalog struct {
	ResourceList []api.Resource
}

type Graph struct {
	Catalog
	ResourceMap map[string]*api.Resource
}

// Check state of parents. Returns true to apply resource, false to cancel
func checkParent(
	ctx context.Context, g *Graph, r api.Resource, p string, afterOk bool,
	afterFail bool, log *logrus.Logger, errChan chan error,
) bool {
	key := r.GetName()
	log.Debugf("Checking if parent %s was applied before applying %s\n", p, key)
	// Fetch parent
	pr, err := g.FetchResource(p, log)
	if err != nil {
		err = r.Fail(log, errors.Errorf(
			"Unable to fetch parent resource %s to check state: %v", r, err,
		))
		if err != nil {
			errChan <- err
		}
		return false
	}
	// Enter inf loop, exit when either completed, failed, or cancelled
afterParentChecks:
	for {
		select {
		case <-ctx.Done():
			err = r.Fail(log, errors.Errorf(
				"Cancelling %s due to context cancellation", key,
			))
			if err != nil {
				errChan <- err
			}
			return false
		default:
			ps := (*pr).GetMetadata().State
			// If waiting for parent to be ok - failed == fail, completed == start
			if afterOk {
				if ps.Failed == true {
					err = r.Fail(log, errors.Errorf(
						"Cancelling %s due to failure of %s", key, p,
					))
					if err != nil {
						errChan <- err
					}
					return false
				}
				if ps.Completed == true {
					log.Debugf("%s: Parent %s completed w/ success\n", key, p)
					break afterParentChecks
				}
			}
			// If waiting for parent to fail - failed == start, completed == cancel
			if afterFail {
				if ps.Failed == true {
					log.Debugf("%s: Parent %s completed w/ failure\n", key, p)
					break afterParentChecks
				}
			}
			if ps.Completed == true {
				log.Debugf("%s: Parent %s completed w/ success, cancelling\n", key, p)
				return false
			}
		}
	}
	time.Sleep(time.Millisecond * 100)
	return true
	//log.Debugf("%s: sleeping till %s\n", key, p)
}

func applyResource(
	ctx context.Context,
	g *Graph, r api.Resource,
	log *logrus.Logger,
	wg *sync.WaitGroup,
	errChan chan error,
) {

	// Defer our clean up
	defer wg.Done()

	// Helper vars
	key := strings.Title(r.GetKind()) + "::" + r.GetMetadata().Name

	// Check if parents applied and done. Returns whether to skip resource or not
	parents := r.GetMetadata().Ordering.AfterOk
	afterOk := true
	afterFail := false
	for _, p := range parents {
		cont := checkParent(ctx, g, r, p, afterOk, afterFail, log, errChan)
		if !cont {
			return
		}
	}
	parents = r.GetMetadata().Ordering.AfterFail
	afterOk = false
	afterFail = true
	for _, p := range parents {
		cont := checkParent(ctx, g, r, p, afterOk, afterFail, log, errChan)
		if !cont {
			return
		}
	}

	log.Infof("%s: start applying\n", key)
	// Apply our state
	if err := r.Apply(ctx, log); err != nil {
		err = r.Fail(log, errors.Errorf(
			"%s: Unable to apply resource: %s", key, err,
		))
		if err != nil {
			errChan <- err
		}
		return
	}

	r.Done()
	log.Infof("%s: applied\n", key)
	return
}

func (g *Graph) Apply(ictx context.Context, log *logrus.Logger) error {

	ctx, cancel := context.WithCancel(ictx)
	log.Debugf("Applying catalog\n")
	log.Debugf("Catalog contents: %+v\n", g.ResourceMap)

	// 	errs.Go(func() error {
	// 		nextRs, err = recursiveApplyResource(g, r, log, errs)
	// 		if err != nil { return err }
	// 		for _, r := range nextRs {
	// 			errs, ctx := errgroup.WithContext(ctx)
	// 			errs.Go(
	// 		return
	// 	}())

	// Spawn each resource to apply in a separate goroutine
	// Each routine is responsible for checking resources it relies on until it's
	// ready to apply
	errChan := make(chan error)
	defer close(errChan)
	resourceApplyWG := &sync.WaitGroup{}
	for _, r := range g.ResourceList {
		resourceApplyWG.Add(1)
		go applyResource(ctx, g, r, log, resourceApplyWG, errChan)
	}

	// Spin up go routine to handle errors from the resource apply go routines
	var errorHandlerWG sync.WaitGroup
	errorHandlerWG.Add(1)
	go func() {
		defer errorHandlerWG.Done()
		for {
			select {
			case err := <-errChan:
				log.Errorf("%v\n", err)
			case <-ctx.Done():
				cancel()
				return
			}
		}
	}()

	// Wait for resource go routines to exit
	resourceApplyWG.Wait()
	// Cancel the context, killing error handler go routine
	cancel()
	// Wait for error handler to exit
	errorHandlerWG.Wait()

	log.Infoln("Applied catalog")
	return nil
}

func NewCatalog(fp string, log *logrus.Logger) (*Catalog, error) {
	log.Debugf("Loading catalog from %s\n", fp)
	data, err := ioutil.ReadFile(fp)
	if err != nil {
		return nil, errors.Errorf("Unable to read %s: %s\n", fp, err)
	}
	c := Catalog{}
	var rl []map[string]interface{}
	if err := json.Unmarshal(data, &rl); err != nil {
		return nil, errors.Errorf(
			"Unable to unmarshal data from %s: %s\n", fp, err,
		)
	}
	for _, elem := range rl {
		var res api.Resource
		resKind, ok := elem["kind"]
		if !ok {
			return nil, errors.Errorf(
				"Unable to find key `kind` in resource %v", elem,
			)
		}
		resKindStr, ok := resKind.(string)
		if !ok {
			return nil, errors.Errorf(
				"Unable to convert value %v to string", resKind,
			)
		}
		switch strings.ToLower(resKindStr) {
		case "file":
			f := &resources.File{}
			// Re-marshal elem into json binary
			b, err := json.Marshal(elem)
			if err != nil {
				return nil, errors.Errorf(
					"Unable to re-marshall map of kind File into json for File::%s: %v",
					resKindStr, err,
				)
			}
			log.Debugf("remarshalled is %s\n", b)
			err = json.Unmarshal(b, f)
			if err != nil {
				return nil, errors.Errorf(
					"Unable to load resource %s of kind File into Resources.File: %v",
					resKindStr, err,
				)
			}
			// init the state struct
			f.Init()
			log.Debugf("Loaded resource %+v\n", *f)
			res = f
		case "exec":
			e := &resources.Exec{}
			// Re-marshal elem into json binary
			b, err := json.Marshal(elem)
			if err != nil {
				return nil, errors.Errorf(
					"Unable to re-marshall map of kind Exec into json for Exec::%s: %v",
					resKindStr, err,
				)
			}
			log.Debugf("remarshalled is %s\n", b)
			err = json.Unmarshal(b, e)
			if err != nil {
				return nil, errors.Errorf(
					"Unable to load resource %s of kind Exec into Resources.Exec: %v",
					resKindStr, err,
				)
			}
			// init the state struct
			e.Init()
			log.Debugf("Loaded resource %+v\n", *e)
			res = e
		default:
			return nil, errors.Errorf(
				"Type %s not supported for resource %v", resKindStr, elem,
			)
		}

		c.ResourceList = append(c.ResourceList, res)
		log.Debugf("Added resource %s::%s to catalog\n", resKindStr, res.GetMetadata().Name)
		log.Debugf("Content %+v\n", res)
	}
	log.Debugf("Successfully parsed catalog from %s\n", fp)
	return &c, nil
}

func (g *Graph) LoadCatalog(c *Catalog, log *logrus.Logger) error {
	g.ResourceList = c.ResourceList
	g.ResourceMap = make(map[string]*api.Resource)
	for index := range c.ResourceList {
		kind := strings.Title(c.ResourceList[index].GetKind())
		name := c.ResourceList[index].GetMetadata().Name
		key := kind + "::" + name
		if _, keyAlreadySet := g.ResourceMap[key]; keyAlreadySet {
			return errors.Errorf(
				"Key %s already set in graph. Check catalog for duplicated resources",
				key,
			)
		}
		g.ResourceMap[key] = &c.ResourceList[index]
		log.Debugf("Added resource %s to resourceMap\n", key)
	}
	return nil
}

func (g *Graph) FetchResource(
	resourceName string, log *logrus.Logger,
) (*api.Resource, error) {
	val, ok := g.ResourceMap[resourceName]
	if !ok {
		return nil, errors.Errorf(
			"Unable to find resource %s in graph", resourceName,
		)
	}
	return val, nil
}
