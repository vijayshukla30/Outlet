package main

import (
	"context"
	"fmt"
	"github.com/micro/go-micro"
	pb "github.com/vijayshukla30/Outlet/consignment-service/proto/consignment"
	vesselProto "github.com/vijayshukla30/Outlet/vessel-service/proto/vessel"
	"log"
)

type repository interface {
	Create(*pb.Consignment) (*pb.Consignment, error)
	GetAll() []*pb.Consignment
}

type Repository struct {
	consignments []*pb.Consignment
}

func (repo *Repository) Create(consignment *pb.Consignment) (*pb.Consignment, error) {
	updated := append(repo.consignments, consignment)
	repo.consignments = updated
	return consignment, nil
}

func (repo *Repository) GetAll() []*pb.Consignment {
	return repo.consignments
}

type service struct {
	repo         repository
	vesselClient vesselProto.VesselServiceClient
}

func (s *service) CreateConsignment(ctx context.Context, req *pb.Consignment, res *pb.Response) (error) {

	vesselResponse, err := s.vesselClient.FindAvailable(context.Background(), &vesselProto.Specification{
		MaxWeight: req.Weight,
		Capacity:  int32(len(req.Containers)),
	})

	log.Printf("Found vessel: %s \n", vesselResponse.Vessel.Name)
	if err != nil {
		return err
	}

	req.VesselId = vesselResponse.Vessel.Id

	consignment, err := s.repo.Create(req)
	if err != nil {
		return err
	}

	res.Created = true
	res.Consignment = consignment
	return nil
}

func (s *service) GetConsignments(ctx context.Context, req *pb.GetRequest, res *pb.Response) (error) {
	consignments := s.repo.GetAll()
	res.Consignments = consignments
	return nil
}

func main() {
	repo := &Repository{}

	srv := micro.NewService(
		micro.Name("consignment.service"),
	)

	//Init will parse the command line flags
	srv.Init()

	vesselClient := vesselProto.NewVesselServiceClient("vessel.service", srv.Client())
	//Register Handler
	pb.RegisterShippingServiceHandler(srv.Server(), &service{repo, vesselClient})

	//Run the Server
	if err := srv.Run(); err != nil {
		fmt.Println(err)
	}
}
