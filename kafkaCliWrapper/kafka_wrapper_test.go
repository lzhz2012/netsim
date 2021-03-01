package kafkaCliWrapper

import (
	"fmt"
	"sync"
	"testing"

	log "github.com/sirupsen/logrus"
)

// TASK_READY_STAUS
var testData string = `{
    " taskId ": "试验任务id",
     "systemType": "WLXWFZ",
     "status": 1,
     "timestamp": 1581576778
}`

func TestSyncProducer(t *testing.T) {

	result := `{"success": true, "status": %d}`
	fmt.Sprintf(result, 0)
	//serverAddress := []string{"kafakaServer:49162", "kafakaServer:49159", "kafakaServer:49160", "kafakaServer:49161"}
	topic := []string{"test"}
	log.Printf("%X", []byte(testData))
	SyncProducer(serverAddress, topic[0], testData)
	//AsyncProducer(testServer)
	var wg = &sync.WaitGroup{}
	wg.Add(1)

	//ClusterConsumer(wg, serverAddress, topic, "group-1")
	Consumer(wg, topic[0], serverAddress)
	wg.Wait()
}
