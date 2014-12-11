package transporter

import (
	"fmt"
	"time"

	"github.com/compose/transporter/pkg/adaptor"
	"github.com/compose/transporter/pkg/events"
	"github.com/compose/transporter/pkg/state"
)

const (
	VERSION = "0.0.1"
)

// A Pipeline is a the end to end description of a transporter data flow.
// including the source, sink, and all the transformers along the way
type Pipeline struct {
	source        *Node
	emitter       events.Emitter
	sessionStore  state.SessionStore
	metricsTicker *time.Ticker
	sessionTicker *time.Ticker
}

// NewDefaultPipeline returns a new Transporter Pipeline with the given node tree, and
// uses the events.HttpPostEmitter to deliver metrics.
// eg.
//   source :=
//   	transporter.NewNode("source", "mongo", adaptor.Config{"uri": "mongodb://localhost/", "namespace": "boom.foo", "debug": false, "tail": true}).
// 	  	Add(transporter.NewNode("out", "file", adaptor.Config{"uri": "stdout://"}))
//   pipeline, err := transporter.NewDefaultPipeline(source, events.Api{Uri: "http://localhost/endpoint"}, 1*time.Second)
//   if err != nil {
// 	  fmt.Println(err)
// 	  os.Exit(1)
//   }
// pipeline.Run()
func NewDefaultPipeline(source *Node, uri, key, pid string, interval time.Duration) (*Pipeline, error) {
	emitter := events.HttpPostEmitter(uri, key, pid)
	sessionStore := state.NewFilestore(key, "/tmp/transporter.state", interval)
	return NewPipeline(source, emitter, interval, sessionStore)
}

// NewPipeline creates a new Transporter Pipeline using the given tree of nodes, and Event Emitter
// eg.
//   source :=
//   	transporter.NewNode("source", "mongo", adaptor.Config{"uri": "mongodb://localhost/", "namespace": "boom.foo", "debug": false, "tail": true}).
// 	  	Add(transporter.NewNode("out", "file", adaptor.Config{"uri": "stdout://"}))
//   pipeline, err := transporter.NewPipeline(source, events.NewNoopEmitter(), 1*time.Second)
//   if err != nil {
// 	  fmt.Println(err)
// 	  os.Exit(1)
//   }
// pipeline.Run()
func NewPipeline(source *Node, emitter events.Emitter, interval time.Duration, sessionStore state.SessionStore) (*Pipeline, error) {
	pipeline := &Pipeline{
		source:        source,
		emitter:       emitter,
		sessionStore:  sessionStore,
		metricsTicker: time.NewTicker(interval),
		sessionTicker: time.NewTicker(10 * time.Second),
	}

	// init the pipeline
	err := pipeline.source.Init(interval)
	if err != nil {
		return pipeline, err
	}

	// init the emitter with the right chan
	pipeline.emitter.Init(source.pipe.Event)

	// start the emitters
	go pipeline.startErrorListener(source.pipe.Err)
	go pipeline.startMetricsGatherer()
	go pipeline.startSessionSaver()
	pipeline.emitter.Start()

	return pipeline, nil
}

func (pipeline *Pipeline) String() string {
	out := pipeline.source.String()
	return out
}

// Stop sends a stop signal to the emitter and all the nodes, whether they are running or not.
// the node's database adaptors are expected to clean up after themselves, and stop will block until
// all nodes have stopped successfully
func (pipeline *Pipeline) Stop() {
	pipeline.source.Stop()
	pipeline.emitter.Stop()
	pipeline.sessionTicker.Stop()
	pipeline.metricsTicker.Stop()
}

// run the pipeline
func (pipeline *Pipeline) Run() error {
	endpoints := pipeline.source.Endpoints()
	// send a boot event
	pipeline.source.pipe.Event <- events.BootEvent(time.Now().Unix(), VERSION, endpoints)

	// start the source
	err := pipeline.source.Start()

	// pipeline has stopped, emit one last round of metrics and send the exit event
	pipeline.emitNodeMetadata(pipeline.fireMetrics)
	pipeline.emitNodeMetadata(pipeline.saveState)
	pipeline.source.pipe.Event <- events.ExitEvent(time.Now().Unix(), VERSION, endpoints)

	// the source has exited, stop all the other nodes
	pipeline.Stop()

	// send a boot event

	return err
}

// start error listener consumes all the events on the pipe's Err channel, and stops the pipeline
// when it receives one
func (pipeline *Pipeline) startErrorListener(cherr chan error) {
	for err := range cherr {
		if aerr, ok := err.(adaptor.Error); ok {
			fmt.Printf("we got an adaptor error, %+v\n", aerr)
			pipeline.source.pipe.Event <- events.ErrorEvent(time.Now().Unix(), aerr.Path, aerr.Record, aerr.Error())
		} else {
			fmt.Printf("Pipeline error %v\nShutting down pipeline\n", err)
			pipeline.Stop()
		}
	}
}

func (pipeline *Pipeline) startMetricsGatherer() {
	for _ = range pipeline.metricsTicker.C {
		pipeline.emitNodeMetadata(pipeline.fireMetrics)
	}
}

func (pipeline *Pipeline) fireMetrics(node *Node) {
	// do something with the node
	pipeline.source.pipe.Event <- events.MetricsEvent(time.Now().Unix(), node.Path(), node.pipe.MessageCount)
}

func (pipeline *Pipeline) startSessionSaver() {
	for _ = range pipeline.sessionTicker.C {
		pipeline.emitNodeMetadata(pipeline.saveState)
	}
}

func (pipeline *Pipeline) saveState(node *Node) {
	if node.pipe.LastKnownState != nil {
		pipeline.sessionStore.Save(node.Path(), node.pipe.LastKnownState)
	}
}

// emit the metrics
func (pipeline *Pipeline) emitNodeMetadata(fn func(*Node)) {

	frontier := make([]*Node, 1)
	frontier[0] = pipeline.source

	for {
		// pop the first item
		node := frontier[0]
		frontier = frontier[1:]

		fn(node)

		// add this nodes children to the frontier
		for _, child := range node.Children {
			frontier = append(frontier, child)
		}

		// if we're empty
		if len(frontier) == 0 {
			break
		}
	}
}
