package common

func IsChanClosed(channel chan bool) bool {
	select {
	case _, ok := <-channel:
		return !ok
	default:
		return false
	}
}
