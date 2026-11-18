package logs

import "time"

const spinnerInterval = 100 * time.Millisecond

var spinnerFrames = [...]string{"|", "/", "-", "\\"}

func (logger *Logger) animateFooter(state *footerState) {
	ticker := time.NewTicker(spinnerInterval)
	defer ticker.Stop()
	defer close(state.done)
	for {
		select {
		case <-state.stop:
			return
		case <-ticker.C:
			logger.mu.Lock()
			if state.stopping || state.err != nil {
				logger.mu.Unlock()
				return
			}
			state.frame = (state.frame + 1) % len(spinnerFrames)
			if err := logger.renderFooter(state); err != nil {
				state.err = err
				logger.mu.Unlock()
				return
			}
			logger.mu.Unlock()
		}
	}
}
