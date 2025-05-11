package streaming

import (
	"context"
	"time"

	"github.com/livekit/protocol/auth"
	"github.com/livekit/protocol/livekit"
	lksdk "github.com/livekit/server-sdk-go/v2"
)

type Streaming struct {
	clientRoom    *lksdk.RoomServiceClient
	clientIngress *lksdk.IngressClient
	config        Config
}
type Config struct {
	Key    string
	Secret string
	Url    string
}

func NewStreaming(cfg Config) *Streaming {
	return &Streaming{
		clientRoom:    lksdk.NewRoomServiceClient(cfg.Url, cfg.Key, cfg.Secret),
		clientIngress: lksdk.NewIngressClient(cfg.Url, cfg.Key, cfg.Secret),
		config:        Config{cfg.Key, cfg.Secret, cfg.Url},
	}
}

func (s Streaming) GetRTMP(roomName, name, identity string) (*livekit.IngressInfo, error) {
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

func (s Streaming) JoinToken(roomName, userIdentity string) (string, error) {
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

func (s Streaming) CreateRoom(roomName string) (*livekit.Room, error) {
	return s.clientRoom.CreateRoom(context.Background(), &livekit.CreateRoomRequest{
		Name: roomName,
	})
}

func (s Streaming) ListRooms() (*livekit.ListRoomsResponse, error) {
	return s.clientRoom.ListRooms(context.Background(), &livekit.ListRoomsRequest{})
}

func (s Streaming) DeleteRoom(id string) {
	s.clientRoom.DeleteRoom(context.Background(), &livekit.DeleteRoomRequest{
		Room: id,
	})
}

func (s Streaming) KickUser(roomName, userIdentity string) {
	s.clientRoom.RemoveParticipant(context.Background(), &livekit.RoomParticipantIdentity{
		Room:     roomName,
		Identity: userIdentity,
	})
}
