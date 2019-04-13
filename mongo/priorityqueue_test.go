package mongo

import (
	"testing"
)

func TestNewPriorityQueue(t *testing.T) {
	queue, _ := NewPriorityQueue("mongodb://127.0.0.1:27017", "ConvertJob", "")
	//queue.AddTask("0B0EM1NfwGeVfMEtCVkRZcWh2QnM", "0B0EM1NfwGeVfMEtCVkRZcWh2QnM_360", 1)
	//queue.AddTask("0B0EwGeVfMEtCVkRZcWh2QnM", "0B0EM1NfwGeVfMEtCVkRZcWh2QnM_480", 1)
	queue.AddTask("0B0EM1NfeVfMEtCVkRZcWh2QnM", "0B0EM1NfwGeVfMEtCVkRZcWh2QnM_586", 1)
	queue.AddTask("0BEM1NwGeVfMEtCVkRZcWh2QnM", "0B0EM1NfwGeVfMEtCVkRZcWh2QM_360", 1)
}
