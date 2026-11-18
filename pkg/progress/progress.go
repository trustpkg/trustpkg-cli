package progress

import "github.com/trustpkg/trustpkg-cli-forge/internal/progress"

type (
	ProgressUpdate   = progress.Update
	ProgressCallback = progress.Callback
)

func NewProgressReporter(totalSteps int, onProgress ProgressCallback) func() {
	return progress.NewReporter(totalSteps, onProgress)
}
