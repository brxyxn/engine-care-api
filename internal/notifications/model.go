package notifications

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type Channel string

const (
	ChannelEmail    Channel = "email"
	ChannelSMS      Channel = "sms"
	ChannelWhatsApp Channel = "whatsapp"
	ChannelPush     Channel = "push"
	ChannelWebhook  Channel = "webhook"
)

type NotificationLog struct {
	bun.BaseModel `bun:"table:notification_logs,alias:nl"`

	ID             uuid.UUID       `bun:"id,pk,default:gen_random_uuid()" json:"id"`
	OrganizationID uuid.UUID       `bun:"organization_id,notnull" json:"organization_id"`
	WorkOrderID    *uuid.UUID      `bun:"work_order_id" json:"work_order_id,omitempty"`
	EventID        *uuid.UUID      `bun:"event_id" json:"event_id,omitempty"`
	CustomerID     *uuid.UUID      `bun:"customer_id" json:"customer_id,omitempty"`
	Channel        Channel         `bun:"channel,type:notify_channel,notnull" json:"channel"`
	Recipient      string          `bun:"recipient,notnull" json:"recipient"`
	TemplateKey    *string         `bun:"template_key" json:"template_key,omitempty"`
	SentAt         time.Time       `bun:"sent_at,notnull,default:now()" json:"sent_at"`
	Meta           json.RawMessage `bun:"meta,type:jsonb,notnull,default:'{}'" json:"meta"`
}
