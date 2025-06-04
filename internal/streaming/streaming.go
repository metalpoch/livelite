package streaming

import (
	"context"
	"time"

	"github.com/livekit/protocol/auth"
	"github.com/livekit/protocol/livekit"
	lksdk "github.com/livekit/server-sdk-go/v2"
)

type Streaming interface {
	GetRTMP(roomName, name, identity string) (*livekit.IngressInfo, error)
	JoinToken(roomName, userIdentity string) (string, error)
	CreateRoom(roomName string) (*livekit.Room, error)
	ListRooms() (*livekit.ListRoomsResponse, error)
	DeleteRoom(id string)
	KickUser(roomName, userIdentity string)
}

type streaming struct {
	clientRoom    *lksdk.RoomServiceClient
	clientIngress *lksdk.IngressClient
	clientEgress  *lksdk.EgressClient

	config Config
}
type Config struct {
	Key    string
	Url    string
	Secret string
}

func NewStreaming(cfg Config) *streaming {
	return &streaming{
		clientRoom:    lksdk.NewRoomServiceClient(cfg.Url, cfg.Key, cfg.Secret),
		clientIngress: lksdk.NewIngressClient(cfg.Url, cfg.Key, cfg.Secret),
		clientEgress:  lksdk.NewEgressClient(cfg.Url, cfg.Key, cfg.Secret),
		config:        cfg,
	}
}

func (s *streaming) GetRTMP(roomName, name, identity string) (*livekit.IngressInfo, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req := &livekit.CreateIngressRequest{
		InputType:           livekit.IngressInput_RTMP_INPUT,
		RoomName:            roomName,
		Name:                name,
		ParticipantIdentity: identity,
	}

	return s.clientIngress.CreateIngress(ctx, req)
}

func (s *streaming) JoinToken(roomName, userIdentity string) (string, error) {
	at := auth.NewAccessToken(s.config.Key, s.config.Secret)
	grant := &auth.VideoGrant{
		RoomJoin: true,
		Room:     roomName,
	}
	at.SetVideoGrant(grant).
		SetIdentity(userIdentity).
		SetValidFor(time.Hour)

	return at.ToJWT()
}

func (s *streaming) CreateRoom(roomName string) (*livekit.Room, error) {
	return s.clientRoom.CreateRoom(context.Background(), &livekit.CreateRoomRequest{
		Name: roomName,
	})
}

func (s *streaming) ListRooms() (*livekit.ListRoomsResponse, error) {
	return s.clientRoom.ListRooms(context.Background(), &livekit.ListRoomsRequest{})
}

func (s *streaming) DeleteRoom(id string) {
	s.clientRoom.DeleteRoom(context.Background(), &livekit.DeleteRoomRequest{
		Room: id,
	})
}

func (s *streaming) KickUser(roomName, userIdentity string) {
	s.clientRoom.RemoveParticipant(context.Background(), &livekit.RoomParticipantIdentity{
		Room:     roomName,
		Identity: userIdentity,
	})
}
