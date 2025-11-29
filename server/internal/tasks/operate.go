package tasks

func GetParentDependency(t *Task) (*[]Task, error) {
	return nil, nil
}

func GetChildDependency(t *Task) (*[]Task, error) {
	return nil, nil
}

func SaveTask(t *Task) error {
	return nil
}

func (t Task) IsReady() (bool, error) {
	return true, nil
}

func (t Task) MakeRunning() (bool, error) {
	return true, nil
}

func (t Task) MakeWaiting() (bool, error) {
	return true, nil
}

func (t Task) MakeDone() (bool, error) {
	return true, nil
}
