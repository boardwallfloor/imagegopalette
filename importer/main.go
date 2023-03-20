package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"

	pb "imagegopalette/importer/pb"

	"google.golang.org/grpc"
)

func main() {
	fmt.Println("Starting importer")
	conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Unable to dial %s", err)
	}
	defer conn.Close()

	client := pb.NewImageServiceClient(conn)
	stream, err := client.SendImage(context.Background())
	if err != nil {
		log.Fatalf("Error sending image : %s", err)
	}

	//  Open image
	image, err := os.Open("main.jpg")
	if err != nil {
		fmt.Printf("Error : %s", err)
	}
	defer image.Close()

	// Create stream
	img := bufio.NewReader(image)

	// Target file
	file, err := os.Create("test.jpg")
	if err != nil {
		log.Fatalf("Failed to create file: %v", err)
	}
	defer file.Close()

	// block function
	waitc := make(chan struct{})

	// Streaming image
	buf := make([]byte, 4096)
	go func() {
		for {
			stat, err := file.Stat()
			if err != nil {
				log.Fatalf("Stat error : %s", err)
			}

			// Stat confirmation
			fmt.Println(stat.Size())
			cmd := exec.Command("ls", "-lh")
			stdout, err := cmd.Output()
			if err != nil {
				fmt.Println(err.Error())
				return
			}
			fmt.Println(string(stdout))
			// time.Sleep(1 * time.Second)
		}
	}()
	for {
		n, err := img.Read(buf)
		if err == io.EOF {
			break
		} else if err != nil {
			log.Fatalf("Read error : %s", err)
		}

		err = stream.Send(&pb.ImageData{Data: buf[:n]})
		if err != nil {
			log.Fatalf("Error sending stream : %s", err)
		}
	}

	go func() {
		for {
			in, err := stream.Recv()
			if err == io.EOF {
				close(waitc)
				return
			}
			_, err = file.Write(in.Data) // write the read data to the file
			if err != nil {
				log.Fatalf("Write error: %s", err)
			}
		}
	}()

	stream.CloseSend()
	<-waitc
}
