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
	StartRoomHLSegress(streamName, filenamePrefix, playlistName, livePlaylistName string) (*livekit.EgressInfo, error)
	DeleteIngress(ingressID string) error
	DeleteEgress(egressID string) error
}

type streaming struct {
	clientRoom    *lksdk.RoomServiceClient
	clientIngress *lksdk.IngressClient
	clientEgress  *lksdk.EgressClient
	s3Upload      *livekit.S3Upload
	config        Config
}

type Config struct {
	Key            string
	Url            string
	Secret         string
	BucketKey      string
	BucketSecret   string
	BucketRegion   string
	BucketEndpoint string
	BucketName     string
}

func NewStreaming(cfg Config) *streaming {
	return &streaming{
		clientEgress:  lksdk.NewEgressClient(cfg.Url, cfg.Key, cfg.Secret),
		clientRoom:    lksdk.NewRoomServiceClient(cfg.Url, cfg.Key, cfg.Secret),
		clientIngress: lksdk.NewIngressClient(cfg.Url, cfg.Key, cfg.Secret),
		s3Upload: &livekit.S3Upload{
			AccessKey: cfg.BucketKey,
			Secret:    cfg.BucketSecret,
			Region:    cfg.BucketRegion,
			Endpoint:  cfg.BucketEndpoint,
			Bucket:    cfg.BucketName,
		},
		config: cfg,
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

func (s *streaming) StartRoomHLSegress(streamName, filenamePrefix, playlistName, livePlaylistName string) (*livekit.EgressInfo, error) {
	req := &livekit.RoomCompositeEgressRequest{
		RoomName: streamName,
		SegmentOutputs: []*livekit.SegmentedFileOutput{
			{
				FilenamePrefix:   filenamePrefix,
				PlaylistName:     playlistName,
				LivePlaylistName: livePlaylistName,
				SegmentDuration:  60,
				Output: &livekit.SegmentedFileOutput_S3{
					S3: s.s3Upload,
				},
			},
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	return s.clientEgress.StartRoomCompositeEgress(ctx, req)
}

func (s *streaming) DeleteIngress(ingressID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_, err := s.clientIngress.DeleteIngress(ctx, &livekit.DeleteIngressRequest{
		IngressId: ingressID,
	})
	return err
}

func (s *streaming) DeleteEgress(egressID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_, err := s.clientEgress.StopEgress(ctx, &livekit.StopEgressRequest{
		EgressId: egressID,
	})
	return err
}
