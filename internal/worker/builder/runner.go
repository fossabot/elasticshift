/*
Copyright 2018 The Elasticshift Authors.
*/
package builder

import (
	"log"
	"os"
	"runtime"
	"sync"

	homedir "github.com/minio/go-homedir"
	"gitlab.com/conspico/elasticshift/api"
	"gitlab.com/conspico/elasticshift/internal/pkg/graph"
	"gitlab.com/conspico/elasticshift/internal/pkg/utils"
)

func (b *builder) build(g *graph.Graph) error {

	wdir := b.f.WorkDir()
	log.Printf("Working directory : %s\n", wdir)

	if wdir != "" {
		expanded, err := homedir.Expand(wdir)
		if err != nil {
			log.Printf("Failed to expand the directory : %v\n", err)
		}

		err = utils.Mkdir(expanded)
		if err != nil {
			log.Printf("Failed to create working directory : %v\n", err)
		}

		err = os.Chdir(expanded)
		if err != nil {
			log.Printf("Failed to change the working directory : %v\n", err)
		}
	}

	log.Println("Working directory = " + utils.GetWD())

	// set the parallel capability
	var parallel int
	nCpu := runtime.NumCPU()
	if nCpu < 2 {
		parallel = 1
	} else {
		parallel = nCpu - 1
	}

	var failed bool

	// walk through the checkpoints and execute them
	// s := ""
	checkpoints := g.Checkpoints()
	for i := 0; i < len(checkpoints); i++ {

		if failed {
			log.Println("Build finished, waiting from shiftserver to receive a halt command..")
			<-b.done
		}

		c := checkpoints[i]
		// s += fmt.Sprintf("(%d) %s\n", i+1, c.node.Name())

		// run block if it is a sequential task
		// If it's FANOUT, spawn the multiple block
		// with in a worker group and wait for it to complete
		edgeSize := len(c.Edges)
		if edgeSize > 0 {

			var errMutex sync.Mutex
			var wg sync.WaitGroup
			parallelCh := make(chan int, parallel)

			for j := 0; j < edgeSize; j++ {

				if failed {
					break
				}

				wg.Add(1)

				go func(n *graph.N) {

					n.Wait()
					b.UpdateBuildGraphToShiftServer(graph.StatusWaiting, n.Name)

					defer wg.Done()
					parallelCh <- 1

					n.Start()
					b.UpdateBuildGraphToShiftServer(graph.StatusRunning, n.Name)

					msg, err := b.invokePlugin(n)

					if err != nil {

						errMutex.Lock()
						defer errMutex.Unlock()

						log.Printf("Plugin error : %v\n", err)
						n.End(graph.StatusFailed, msg)
						b.UpdateBuildGraphToShiftServer(graph.StatusFailed, n.Name)

						failed = true
					} else {

						if n.Status != graph.StatusFailed {
							n.End(graph.StatusSuccess, "")
							b.UpdateBuildGraphToShiftServer(graph.StatusSuccess, n.Name)
						}
					}

					<-parallelCh

				}(c.Edges[j])
			}

			// wait until all the parallel tasks are finished
			wg.Wait()

		} else {

			c.Node.Start()
			b.UpdateBuildGraphToShiftServer(graph.StatusRunning, c.Node.Name)

			// sequential checkpoint execution
			msg, err := b.invokePlugin(c.Node)
			if err != nil {
				c.Node.End(graph.StatusFailed, msg)
				log.Printf("Plugin error : %v\n", err)
				b.UpdateBuildGraphToShiftServer(graph.StatusFailed, c.Node.Name)

				failed = true
			} else {

				c.Node.End(graph.StatusSuccess, "")
				b.UpdateBuildGraphToShiftServer(graph.StatusSuccess, c.Node.Name)
			}
		}
	}

	return nil
}

func (b *builder) UpdateBuildGraphToShiftServer(status, checkpoint string) {

	if graph.StatusFailed == status || (graph.END == checkpoint && graph.StatusSuccess == status) {

		log.Println("Saving cache.")

		b.saveCache()

		log.Println("Finished saving the cache")
	}

	gph, err := b.g.JSON()
	if err != nil {
		log.Printf("Eror when contructing status graph: %v", err)
	}

	req := &api.UpdateBuildStatusReq{}
	req.BuildId = b.config.BuildID
	req.Graph = gph
	req.Status = status
	req.Checkpoint = checkpoint

	if b.shiftclient != nil {
		_, err = b.shiftclient.UpdateBuildStatus(b.ctx, req)
		if err != nil {
			log.Printf("Failed to update buld graph: %v", err)
		}
	}
}
