package main

import "github.com/fluent/fluent-bit-go/output"
import (
  "github.com/ugorji/go/codec"
  "github.com/Shopify/sarama"
  "fmt"
  "unsafe"
  "C"
  "reflect"
)

var brokerList []string = []string{"kafka-0.broker.kafka.svc.cluster.local:9092","kafka-1.broker.kafka.svc.cluster.local:9092","kafka-2.broker.kafka.svc.cluster.local:9092"}
var producer sarama.SyncProducer

//export FLBPluginInit
func FLBPluginInit(ctx unsafe.Pointer) int {
  var err error
  producer, err = sarama.NewSyncProducer(brokerList, nil)

  if err != nil {
    fmt.Printf("Failed to start Sarama producer: %v\n", err)
    return output.FLB_ERROR
  }

  fmt.Printf("Up and running")

  return output.FLBPluginRegister(ctx, "out_kafka", "out_kafka GO!")
}

//export FLBPluginFlush
func FLBPluginFlush(data unsafe.Pointer, length C.int, tag *C.char) int {
  var count int
  var h codec.Handle = new(codec.MsgpackHandle)
  var b []byte
  var m interface{}
  var err error

  fmt.Printf("At flush")

  b = C.GoBytes(data, length)
  dec := codec.NewDecoderBytes(b, h)

  // Iterate the original MessagePack array
  count = 0
  for {
    // Decode the entry
    err = dec.Decode(&m)
    if err != nil {
      break
    }

    // Get a slice and their two entries: timestamp and map
    slice := reflect.ValueOf(m)
    timestamp := slice.Index(0)
    data := slice.Index(1)

    // Convert slice data to a real map and iterate
    map_data := data.Interface().(map[interface{}] interface{})
    fmt.Printf("[%d] %s: [%d, {", count, C.GoString(tag), timestamp)
    for k, v := range map_data {
      fmt.Printf("\"%s\": %v, ", k, v)
    }
    fmt.Printf("}\n")
    count++
  }

  // Return options:
  //
  // output.FLB_OK    = data have been processed.
  // output.FLB_ERROR = unrecoverable error, do not try this again.
  // output.FLB_RETRY = retry to flush later.
  return output.FLB_OK
}

//export FLBPluginExit
func FLBPluginExit() int {
  return 0
}

func main() {
}
