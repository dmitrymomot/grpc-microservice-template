package jobs

import (
	"github.com/gocraft/work"
)

// QueueSendEmail ...
const QueueSendEmail string = "send_email"

// SendEmail job handler
func (c *Worker) SendEmail(job *work.Job) error {
	return nil
}
