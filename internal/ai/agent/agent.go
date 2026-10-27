package agent

import "context"

type Agent struct {
	role     string
	maxTurns int
}

func NewAgent(role string, maxTurns int) Agent {
	return Agent{
		role:     role,
		maxTurns: maxTurns,
	}
}

func (a *Agent) Run(ctx context.Context, history []Message, eventsChan chan<- Events) ([]Message, error) {
	for turn := 0; turn < a.maxTurns; turn++ {
		
	}

	return history, nil
}