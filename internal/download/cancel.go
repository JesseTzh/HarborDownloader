package download

import "context"

func IsCancelled(err error) bool {
	return err != nil && (err == context.Canceled || context.Canceled.Error() == err.Error())
}
