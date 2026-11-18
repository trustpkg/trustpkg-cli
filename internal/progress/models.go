package progress

type Update struct {
	CurrentStep int
	TotalSteps  int
}

type Callback func(Update)
