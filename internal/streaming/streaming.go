package streaming

import (
	"context"
	"fmt"
	"time"

	"github.com/livekit/protocol/auth"
	"github.com/livekit/protocol/livekit"
	lksdk "github.com/livekit/server-sdk-go/v2"
)

type Streaming struct {
	clientRoom    *lksdk.RoomServiceClient
	clientIngress *lksdk.IngressClient
	clientEgress  *lksdk.EgressClient

	config Config
}
type Config struct {
	Key              string
	Secret           string
	Url              string
	StorageURL       string
	StorageKey       string
	StorageSecret    string
	StorageBucket    string
	StorageBucketURL string
}

func NewStreaming(cfg Config) *Streaming {
	return &Streaming{
		clientRoom:    lksdk.NewRoomServiceClient(cfg.Url, cfg.Key, cfg.Secret),
		clientIngress: lksdk.NewIngressClient(cfg.Url, cfg.Key, cfg.Secret),
		clientEgress:  lksdk.NewEgressClient(cfg.Url, cfg.Key, cfg.Secret),
		config:        cfg,
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

func (s Streaming) StreamingThumbnails(roomName, identity string) (*livekit.EgressInfo, error) {
	req := &livekit.ParticipantEgressRequest{
		RoomName:    roomName,
		Identity:    identity,
		ScreenShare: false,
		Options: &livekit.ParticipantEgressRequest_Advanced{
			Advanced: &livekit.EncodingOptions{
				Width:            1280,
				Height:           720,
				Framerate:        30,
				AudioCodec:       livekit.AudioCodec_AAC,
				AudioBitrate:     128,
				VideoCodec:       livekit.VideoCodec_H264_HIGH,
				VideoBitrate:     5000,
				KeyFrameInterval: 2,
			},
		},
		StreamOutputs: []*livekit.StreamOutput{{
			Protocol: livekit.StreamProtocol_SRT,
			Urls:     []string{"srt://localhost:9999"},
		}},
		ImageOutputs: []*livekit.ImageOutput{{
			CaptureInterval: 5,
			Width:           1280,
			Height:          720,
			FilenamePrefix:  fmt.Sprintf("streaming/thumbnail/%s/%s", roomName, identity),
			FilenameSuffix:  livekit.ImageFileSuffix_IMAGE_SUFFIX_TIMESTAMP,
			DisableManifest: true,
			Output: &livekit.ImageOutput_S3{
				S3: &livekit.S3Upload{
					Bucket:    s.config.StorageBucket,
					AccessKey: s.config.StorageKey,
					Secret:    s.config.StorageSecret,
					Endpoint:  s.config.StorageURL,
				},
			},
		}},
	}

	return s.clientEgress.StartParticipantEgress(context.Background(), req)
}
