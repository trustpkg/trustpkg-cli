package progress

func NewReporter(totalSteps int, onProgress Callback) func() {
	if totalSteps <= 0 {
		panic("progress: total steps must be positive")
	}

	completedSteps := 0

	return func() {
		if completedSteps >= totalSteps {
			panic("progress: all steps already completed")
		}
		completedSteps++
		if onProgress != nil {
			onProgress(Update{CurrentStep: completedSteps, TotalSteps: totalSteps})
		}
	}
}
