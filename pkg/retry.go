package pkg

import "time"

func Retry(fn func() error, attempts int, waitInterval time.Duration) (err error) {

	for attempts > 0 {

		if err = fn(); err != nil {
			attempts--
			time.Sleep(waitInterval)
			continue
		}

		return nil

	}

	return err

}
