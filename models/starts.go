package models

import (
	"context"
	"database/sql/driver"
	"encoding/json"

	"github.com/jmoiron/sqlx"
	"github.com/nyaruka/gocommon/urns"
	"github.com/nyaruka/goflow/utils/uuids"
	"github.com/nyaruka/null"
	"github.com/pkg/errors"
)

// StartID is our type for flow start idst
type StartID null.Int

// NilStartID is our constant for a nil start id
var NilStartID = StartID(0)

type StartType string

const (
	StartTypeManual     = StartType("M")
	StartTypeAPI        = StartType("A")
	StartTypeFlowAction = StartType("F")
	StartTypeTrigger    = StartType("T")
)

// RestartParticipants is our type for the bool of restarting participatants
type RestartParticipants bool

const DoRestartParticipants = RestartParticipants(true)
const DontRestartParticipants = RestartParticipants(false)

// IncludeActive is our type for the bool of whether to include active contacts
type IncludeActive bool

const DoIncludeActive = IncludeActive(true)
const DontIncludeActive = IncludeActive(false)

// MarkStartComplete sets the status for the passed in flow start
func MarkStartComplete(ctx context.Context, db *sqlx.DB, startID StartID) error {
	_, err := db.Exec("UPDATE flows_flowstart SET status = 'C', modified_on = NOW() WHERE id = $1", startID)
	if err != nil {
		return errors.Wrapf(err, "error setting start as complete")
	}
	return nil
}

// MarkStartStarted sets the status for the passed in flow start to S and updates the contact count on it
func MarkStartStarted(ctx context.Context, db *sqlx.DB, startID StartID, contactCount int) error {
	_, err := db.Exec("UPDATE flows_flowstart SET status = 'S', contact_count = $2, modified_on = NOW() WHERE id = $1", startID, contactCount)
	if err != nil {
		return errors.Wrapf(err, "error setting start as started")
	}
	return nil

}

// FlowStartBatch represents a single flow batch that needs to be started
type FlowStartBatch struct {
	b struct {
		StartID    StartID     `json:"start_id"`
		OrgID      OrgID       `json:"org_id"`
		FlowID     FlowID      `json:"flow_id"`
		FlowType   FlowType    `json:"flow_type"`
		ContactIDs []ContactID `json:"contact_ids"`

		ParentSummary null.JSON `json:"parent_summary,omitempty"`
		Extra         null.JSON `json:"extra,omitempty"`

		RestartParticipants RestartParticipants `json:"restart_participants"`
		IncludeActive       IncludeActive       `json:"include_active"`

		IsLast bool `json:"is_last,omitempty"`
	}
}

func (b *FlowStartBatch) StartID() StartID                         { return b.b.StartID }
func (b *FlowStartBatch) OrgID() OrgID                             { return b.b.OrgID }
func (b *FlowStartBatch) FlowID() FlowID                           { return b.b.FlowID }
func (b *FlowStartBatch) ContactIDs() []ContactID                  { return b.b.ContactIDs }
func (b *FlowStartBatch) RestartParticipants() RestartParticipants { return b.b.RestartParticipants }
func (b *FlowStartBatch) IncludeActive() IncludeActive             { return b.b.IncludeActive }
func (b *FlowStartBatch) IsLast() bool                             { return b.b.IsLast }
func (b *FlowStartBatch) SetIsLast(last bool)                      { b.b.IsLast = last }

func (b *FlowStartBatch) ParentSummary() json.RawMessage { return json.RawMessage(b.b.ParentSummary) }
func (b *FlowStartBatch) Extra() json.RawMessage         { return json.RawMessage(b.b.Extra) }

func (b *FlowStartBatch) MarshalJSON() ([]byte, error)    { return json.Marshal(b.b) }
func (b *FlowStartBatch) UnmarshalJSON(data []byte) error { return json.Unmarshal(data, &b.b) }

// FlowStart represents the top level flow start in our system
type FlowStart struct {
	s struct {
		ID        StartID    `json:"start_id"   db:"id"`
		UUID      uuids.UUID `                  db:"uuid"`
		StartType StartType  `                  db:"start_type"`
		OrgID     OrgID      `json:"org_id"     db:"org_id"`
		FlowID    FlowID     `json:"flow_id"    db:"flow_id"`
		FlowType  FlowType   `json:"flow_type"`

		GroupIDs   []GroupID   `json:"group_ids,omitempty"`
		ContactIDs []ContactID `json:"contact_ids,omitempty"`
		URNs       []urns.URN  `json:"urns,omitempty"`
		Query      null.String `json:"query,omitempty"        db:"query"`

		CreateContact bool `json:"create_contact"`

		RestartParticipants RestartParticipants `json:"restart_participants" db:"restart_participants"`
		IncludeActive       IncludeActive       `json:"include_active"       db:"include_active"`

		Extra         null.JSON `json:"extra,omitempty"          db:"extra"`
		ParentSummary null.JSON `json:"parent_summary,omitempty" db:"parent_summary"`
	}
}

func (s *FlowStart) ID() StartID        { return s.s.ID }
func (s *FlowStart) OrgID() OrgID       { return s.s.OrgID }
func (s *FlowStart) FlowID() FlowID     { return s.s.FlowID }
func (s *FlowStart) FlowType() FlowType { return s.s.FlowType }

func (s *FlowStart) GroupIDs() []GroupID { return s.s.GroupIDs }
func (s *FlowStart) WithGroupIDs(groupIDs []GroupID) *FlowStart {
	s.s.GroupIDs = groupIDs
	return s
}

func (s *FlowStart) ContactIDs() []ContactID { return s.s.ContactIDs }
func (s *FlowStart) WithContactIDs(contactIDs []ContactID) *FlowStart {
	s.s.ContactIDs = contactIDs
	return s
}

func (s *FlowStart) URNs() []urns.URN { return s.s.URNs }
func (s *FlowStart) WithURNs(us []urns.URN) *FlowStart {
	s.s.URNs = us
	return s
}

func (s *FlowStart) Query() string { return string(s.s.Query) }
func (s *FlowStart) WithQuery(query string) *FlowStart {
	s.s.Query = null.String(query)
	return s
}

func (s *FlowStart) RestartParticipants() RestartParticipants { return s.s.RestartParticipants }
func (s *FlowStart) IncludeActive() IncludeActive             { return s.s.IncludeActive }

func (s *FlowStart) CreateContact() bool { return s.s.CreateContact }
func (s *FlowStart) WithCreateContact(create bool) *FlowStart {
	s.s.CreateContact = create
	return s
}

func (s *FlowStart) ParentSummary() json.RawMessage { return json.RawMessage(s.s.ParentSummary) }
func (s *FlowStart) WithParentSummary(sum json.RawMessage) *FlowStart {
	s.s.ParentSummary = null.JSON(sum)
	return s
}

func (s *FlowStart) Extra() json.RawMessage { return json.RawMessage(s.s.Extra) }
func (s *FlowStart) WithExtra(extra json.RawMessage) *FlowStart {
	s.s.Extra = null.JSON(extra)
	return s
}

func (s *FlowStart) MarshalJSON() ([]byte, error)    { return json.Marshal(s.s) }
func (s *FlowStart) UnmarshalJSON(data []byte) error { return json.Unmarshal(data, &s.s) }

// GetFlowStartAttributes gets the basic attributes for the passed in start id, this includes ONLY its id, uuid, flow_id and extra
func GetFlowStartAttributes(ctx context.Context, db Queryer, startID StartID) (*FlowStart, error) {
	start := &FlowStart{}
	err := db.GetContext(ctx, &start.s, `SELECT id, uuid, flow_id, extra, parent_summary FROM flows_flowstart WHERE id = $1`, startID)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to load start attributes for id: %d", startID)
	}
	return start, nil
}

// NewFlowStart creates a new flow start objects for the passed in parameters
func NewFlowStart(orgID OrgID, startType StartType, flowType FlowType, flowID FlowID, restartParticipants RestartParticipants, includeActive IncludeActive) *FlowStart {
	s := &FlowStart{}
	s.s.UUID = uuids.New()
	s.s.OrgID = orgID
	s.s.StartType = startType
	s.s.FlowType = flowType
	s.s.FlowID = flowID
	s.s.RestartParticipants = restartParticipants
	s.s.IncludeActive = includeActive

	return s
}

type startContact struct {
	StartID   StartID   `db:"start_id"`
	ContactID ContactID `db:"contact_id"`
}

type startGroup struct {
	StartID StartID `db:"start_id"`
	GroupID GroupID `db:"contactgroup_id"`
}

// InsertFlowStarts inserts all the passed in starts
func InsertFlowStarts(ctx context.Context, db Queryer, starts []*FlowStart) error {
	is := make([]interface{}, len(starts))
	for i, s := range starts {
		// populate UUID if needed
		if s.s.UUID == "" {
			s.s.UUID = uuids.New()
		}

		is[i] = &s.s
	}

	// insert our starts
	err := BulkSQL(ctx, "inserting flow start", db, insertStartSQL, is)
	if err != nil {
		return errors.Wrapf(err, "error inserting flow starts")
	}

	// build up all our contact associations
	contacts := make([]interface{}, 0, len(starts))
	for _, start := range starts {
		for _, contactID := range start.ContactIDs() {
			contacts = append(contacts, &startContact{
				StartID:   start.ID(),
				ContactID: contactID,
			})
		}
	}

	// insert our contacts
	err = BulkSQL(ctx, "inserting flow start contacts", db, insertStartContactsSQL, contacts)
	if err != nil {
		return errors.Wrapf(err, "error inserting flow start contacts for flow")
	}

	// build up all our group associations
	groups := make([]interface{}, 0, len(starts))
	for _, start := range starts {
		for _, groupID := range start.GroupIDs() {
			groups = append(groups, &startGroup{
				StartID: start.ID(),
				GroupID: groupID,
			})
		}
	}

	// insert our groups
	err = BulkSQL(ctx, "inserting flow start groups", db, insertStartGroupsSQL, groups)
	if err != nil {
		return errors.Wrapf(err, "error inserting flow start groups for flow")
	}

	return nil
}

const insertStartSQL = `
INSERT INTO
	flows_flowstart(uuid,  org_id,  flow_id,  start_type,  created_on,  modified_on,  restart_participants,  include_active,  query,  status, extra,  parent_summary)
			 VALUES(:uuid, :org_id, :flow_id, :start_type, NOW(),       NOW(),        :restart_participants, :include_active, :query, 'P',    :extra, :parent_summary)
RETURNING
	id
`

const insertStartContactsSQL = `
INSERT INTO
	flows_flowstart_contacts( flowstart_id,  contact_id)
	                  VALUES(:start_id,     :contact_id)
`

const insertStartGroupsSQL = `
INSERT INTO
	flows_flowstart_groups( flowstart_id,  contactgroup_id)
	                VALUES(:start_id,     :contactgroup_id)
`

// CreateBatch creates a batch for this start using the passed in contact ids
func (s *FlowStart) CreateBatch(contactIDs []ContactID) *FlowStartBatch {
	b := &FlowStartBatch{}
	b.b.StartID = s.ID()
	b.b.OrgID = s.OrgID()
	b.b.FlowID = s.FlowID()
	b.b.FlowType = s.FlowType()
	b.b.ContactIDs = contactIDs
	b.b.RestartParticipants = s.RestartParticipants()
	b.b.IncludeActive = s.IncludeActive()
	b.b.ParentSummary = null.JSON(s.ParentSummary())
	b.b.Extra = null.JSON(s.Extra())
	return b
}

// MarshalJSON marshals into JSON. 0 values will become null
func (i StartID) MarshalJSON() ([]byte, error) {
	return null.Int(i).MarshalJSON()
}

// UnmarshalJSON unmarshals from JSON. null values become 0
func (i *StartID) UnmarshalJSON(b []byte) error {
	return null.UnmarshalInt(b, (*null.Int)(i))
}

// Value returns the db value, null is returned for 0
func (i StartID) Value() (driver.Value, error) {
	return null.Int(i).Value()
}

// Scan scans from the db value. null values become 0
func (i *StartID) Scan(value interface{}) error {
	return null.ScanInt(value, (*null.Int)(i))
}
