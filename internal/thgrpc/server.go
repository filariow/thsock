package thgrpc

import (
	"context"
	"log"

	"github.com/filariow/go-dht"
	"github.com/filariow/thsock/pkg/thprotos"
)

func New() thprotos.TempHumSvcServer {
	return &thGrpcServer{}
}

type thGrpcServer struct {
	thprotos.UnimplementedTempHumSvcServer
}

func (s *thGrpcServer) ReadTempHum(_ context.Context, _ *thprotos.ReadTempHumRequest) (*thprotos.ReadTempHumReply, error) {
	log.Println("Read TempHum request received")

	// sensorType := dht.DHT11
	// sensorType := dht.AM2302
	sensorType := dht.DHT12
	// Read DHT11 sensor data from specific pin, retrying 10 times in case of failure.
	pin := 17
	temperature, humidity, retried, err :=
		dht.ReadDHTxxWithRetry(sensorType, pin, false, 10)
	if err != nil {
		log.Fatal(err)
	}
	// print temperature and humidity
	log.Printf("Sensor = %v: Temperature = %v*C, Humidity = %v%% (retried %d times)",
		sensorType, temperature, humidity, retried)

	return &thprotos.ReadTempHumReply{
		Temperature: float64(temperature),
		Humidity:    float64(humidity),
	}, nil
}
