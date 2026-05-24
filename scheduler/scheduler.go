package scheduler

import (
	"cube/node"
	"cube/task"
)

type Scheduler interface {
	SelectCandidateNodes(t *task.Task, nodes []*node.Node) []*node.Node
	Score(t task.Task, nodes []*node.Node) map[string]float64
	Pick(scores map[string]float64, candidate *[]node.Node) *node.Node
}

type RoundRobin struct {
	Name       string
	LastWorker int
}

func (r *RoundRobin) SelectCandidateNodes(t *task.Task, nodes []*node.Node) []*node.Node {
	return nodes
}

func (r *RoundRobin) Score(t task.Task, nodes []*node.Node) map[string]float64 {
	nodeScores := make(map[string]float64)

	var newWorker int

	if r.LastWorker+1 < len(nodes) {
		newWorker = r.LastWorker + 1
		r.LastWorker++
	} else {
		newWorker = 0
		r.LastWorker = 0
	}

	for idx, nd := range nodes {
		if idx == newWorker {
			nodeScores[nd.Name] = 0.1
		} else {
			nodeScores[nd.Name] = 1.0
		}
	}

	return nodeScores
}

func (r *RoundRobin) Pick(scores map[string]float64, candidates []*node.Node) *node.Node {
	var bestNode *node.Node
	var lowestScore float64
	for inx, nd := range candidates {
		if inx == 0 {
			bestNode = nd
			lowestScore = scores[nd.Name]
		}
		if scores[nd.Name] < lowestScore {
			lowestScore = scores[nd.Name]
			bestNode = nd
		}
	}
	return bestNode
}
