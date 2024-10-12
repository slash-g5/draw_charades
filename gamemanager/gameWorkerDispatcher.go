package gamemanager

import (
	"multiplayer_game/dto"
	"multiplayer_game/service/common/gamedata"
)

type Dispatcher struct {
	GameRCWorkerPool chan chan dto.GameScheduledEvent
}

func NewDispatcher(numWorkers int) *Dispatcher {
	return &Dispatcher{GameRCWorkerPool: make(chan chan dto.GameScheduledEvent, numWorkers)}
}

func (d *Dispatcher) Run(numWorkers int, numChannels int) {
	// Initialise all the workers
	for i := 0; i < numWorkers; i++ {
		gameRCWorker := &GameRCWorker{GameRCWorkerPool: d.GameRCWorkerPool,
			RoundChangeChan: make(chan dto.GameScheduledEvent, numChannels),
			Quit:            gamedata.RoundChangeQuitChan}
		go gameRCWorker.Start()
	}
	// For start action, assuming one worker is fine
	gameStartWorker := &GameStartWorker{gamedata.GameStartChan, gamedata.QuitGameStartChan}
	go gameStartWorker.Start()
	d.Dispatch()
}

func (d *Dispatcher) Dispatch() {
	d.ListenRCEvent()
}

func (d *Dispatcher) ListenRCEvent() {
	for {
		event := <-gamedata.RoundChangeChan
		go func(dto.GameScheduledEvent) {
			workerChan := <-d.GameRCWorkerPool
			workerChan <- event
		}(event)
	}
}
