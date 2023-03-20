package main

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	pb "imagegopalette/converter/pb"
	"io"
	"log"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type imageService struct {
	pb.UnimplementedImageServiceServer
}

func (s *imageService) SendImage(stream pb.ImageService_SendImageServer) error {
	for {
		in, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}

		// do image conversion here
		img, _, err := image.Decode(bytes.NewReader(in.Data))
		if err != nil {
			log.Fatalf("Failed to decode chunks, %v", err)
		}

		bounds := img.Bounds()
		outputImg := image.NewRGBA(bounds)
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
				oldColor := img.At(x, y)
				r, g, b, _ := oldColor.RGBA()

				// Calculate the new sepia color values
				tr := 0.393*float64(r) + 0.769*float64(g) + 0.189*float64(b)
				tg := 0.349*float64(r) + 0.686*float64(g) + 0.168*float64(b)
				tb := 0.272*float64(r) + 0.534*float64(g) + 0.131*float64(b)

				// Set the new sepia color for the pixel
				newColor := color.RGBA64{uint16(tr), uint16(tg), uint16(tb), 0xffff}
				outputImg.Set(x, y, newColor)
			}
		}

		var buf bytes.Buffer
		err = jpeg.Encode(&buf, img, nil)
		if err != nil {
			panic(err)
		}
		imageBytes := buf.Bytes()

		err = stream.Send(&pb.ImageData{Data: imageBytes})
		if err != nil {
			return err
		}

	}

}
func main() {
	fmt.Println("Starting converter")
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	pb.RegisterImageServiceServer(s, &imageService{})
	reflection.Register(s)

	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)

	}
}
