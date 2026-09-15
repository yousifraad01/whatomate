package handlers

// SpawnForTest exposes spawn so external tests can exercise background-task
// tracking without going through a webhook.
func (a *App) SpawnForTest(fn func()) {
	a.spawn(fn)
}
