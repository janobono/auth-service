package server

import (
	"context"
	"errors"
	"github.com/jackc/pgx/v5"
	"github.com/janobono/auth-service/api"
	"github.com/janobono/auth-service/internal/config"
	"github.com/janobono/auth-service/internal/db"
	"github.com/janobono/auth-service/internal/db/repository"
	"github.com/janobono/auth-service/internal/service"
	"github.com/janobono/auth-service/pkg/auth"
	"github.com/janobono/auth-service/pkg/util"
	"google.golang.org/grpc"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
)

type Server struct {
	config *config.Config
}

func New(config *config.Config) *Server {
	return &Server{config}
}

func (s *Server) Start() {
	jwtToken, err := util.NewJwtToken(&util.JwtConfigProperties{
		Expiration: int64(s.config.AppConfig.TokenExpiresIn * 60),
		Issuer:     s.config.AppConfig.TokenIssuer,
	})

	if err != nil {
		log.Fatalf("failed to init jwt token: %v", err)
	}

	dataSource := db.NewDataSource(s.config.DbConfig)
	defer dataSource.Close()

	authorities := initDefaultAuthorities(dataSource, []string{
		"manager",
		"admin",
	})
	initDefaultUser(dataSource, authorities)

	lis, err := net.Listen("tcp", s.config.ServerConfig.Address)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	decoder := service.NewUserDetailDecoder(dataSource, jwtToken)
	interceptor := auth.NewBearerTokenInterceptor(decoder).CheckToken(&[]auth.SecuredMethod{
		auth.SecuredMethod{
			Method:      "/authservice.UserService/SearchUsers",
			Authorities: []string{"manager", "admin"},
		},
		auth.SecuredMethod{
			Method:      "/authservice.UserService/AddUser",
			Authorities: []string{"admin"},
		},
		auth.SecuredMethod{
			Method:      "/authservice.UserService/GetUser",
			Authorities: []string{"manager", "admin"},
		},
		auth.SecuredMethod{
			Method:      "/authservice.UserService/SetUser",
			Authorities: []string{"admin"},
		},
		auth.SecuredMethod{
			Method:      "/authservice.UserService/DeleteUser",
			Authorities: []string{"admin"},
		},
	})

	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(interceptor),
	)

	api.RegisterAuthServiceServer(grpcServer, service.NewAuthService(dataSource, jwtToken))
	api.RegisterUserServiceServer(grpcServer, service.NewUserService(dataSource))

	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()
	log.Printf("gRPC server started on port %s", s.config.ServerConfig.Address)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	<-stop
	log.Println("Shutting down gRPC server...")
	grpcServer.GracefulStop()
	log.Println("gRPC server stopped gracefully")
}

func initDefaultAuthorities(dataSource *db.DataSource, defaultAuthorities []string) *[]repository.SaAuthority {
	authorities, err := dataSource.Queries.GetAuthorities(context.Background())
	if err != nil {
		log.Fatalf("failed to init authorities: %v", err)
	}

	var saAuthorities []repository.SaAuthority

	if len(authorities) == 0 {
		for _, authority := range defaultAuthorities {
			saAuthority, err := dataSource.Queries.AddAuthority(context.Background(), authority)
			if err != nil {
				log.Fatalf("failed to add authority: %v", err)
			}
			saAuthorities = append(saAuthorities, saAuthority)
		}
	}
	return &saAuthorities
}

func initDefaultUser(dataSource *db.DataSource, authorities *[]repository.SaAuthority) {
	_, err := dataSource.Queries.GetUserByEmail(context.Background(), "simple@auth.org")
	if errors.Is(err, pgx.ErrNoRows) {
		user, err := dataSource.Queries.AddUser(context.Background(), repository.AddUserParams{
			Email:     "simple@auth.org",
			Password:  "$2a$10$Ae8eRIJVL2n8l3UCZUx6H.6obQ0nN8JIDim4gGa/eaA5gv3zLnIzu",
			FirstName: "Simple",
			LastName:  "Auth",
			Confirmed: true,
			Enabled:   true,
		})

		if err != nil {
			log.Fatalf("failed to init default user: %v", err)
		}

		for _, authority := range *authorities {
			err := dataSource.Queries.AddUserAuthority(context.Background(), repository.AddUserAuthorityParams{
				UserID:      user.ID,
				AuthorityID: authority.ID,
			})

			if err != nil {
				log.Fatalf("failed to init default user authority: %v", err)
			}
		}
		return
	}
	if err != nil {
		log.Fatalf("failed to check default user: %v", err)
	}
}
