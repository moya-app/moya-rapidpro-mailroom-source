package handlers

import (
	"context"

	"github.com/gomodule/redigo/redis"
	"github.com/jmoiron/sqlx"
	"github.com/juju/errors"
	"github.com/nyaruka/goflow/flows"
	"github.com/nyaruka/goflow/flows/events"
	"github.com/nyaruka/mailroom/models"
	"github.com/sirupsen/logrus"
)

func init() {
	models.RegisterEventHandler(events.TypeContactGroupsRemoved, handleContactGroupsRemoved)
	models.RegisterEventHandler(events.TypeContactGroupsAdded, handleContactGroupsAdded)
}

// ContactGroupsChangedHook is our hook for all group changes
type ContactGroupsChangedHook struct{}

var contactGroupsChangedHook = &ContactGroupsChangedHook{}

// Apply squashes and delete all our contact groups
func (h *ContactGroupsChangedHook) Apply(ctx context.Context, tx *sqlx.Tx, rp *redis.Pool, orgID models.OrgID, sessions map[*models.Session][]interface{}) error {
	// build up our list of all adds and removes
	adds := make([]interface{}, 0, len(sessions))
	removes := make([]interface{}, 0, len(sessions))

	// we remove from our groups at once, build up our list
	for _, events := range sessions {
		// we use these sets to track what our final add or remove should be
		seenAdds := make(map[models.GroupID]*GroupAdd)
		seenRemoves := make(map[models.GroupID]*GroupRemove)

		for _, e := range events {
			switch event := e.(type) {
			case *GroupAdd:
				seenAdds[event.GroupID] = event
				delete(seenRemoves, event.GroupID)
			case *GroupRemove:
				seenRemoves[event.GroupID] = event
				delete(seenAdds, event.GroupID)
			}
		}

		for _, add := range seenAdds {
			adds = append(adds, add)
		}

		for _, remove := range seenRemoves {
			removes = append(removes, remove)
		}
	}

	// do our updates
	if len(adds) > 0 {
		err := models.BulkInsert(ctx, tx, addContactsToGroupsSQL, adds)
		if err != nil {
			return errors.Annotatef(err, "error adding contacts to groups")
		}
	}
	if len(removes) > 0 {
		err := models.BulkInsert(ctx, tx, removeContactsFromGroupsSQL, removes)
		if err != nil {
			return errors.Annotatef(err, "error removing contacts from groups")
		}
	}

	return nil
}

// handleContactGroupsRemoved is called when a group is removed from a contact
func handleContactGroupsRemoved(ctx context.Context, tx *sqlx.Tx, rp *redis.Pool, session *models.Session, e flows.Event) error {
	event := e.(*events.ContactGroupsRemovedEvent)
	logrus.WithFields(logrus.Fields{
		"contact_uuid": session.ContactUUID(),
		"groups":       event.Groups,
	}).Debug("removing contact from groups")

	// remove each of our groups
	for _, g := range event.Groups {
		// look up our group id
		group := session.Org().GroupByUUID(g.UUID)
		if group == nil {
			logrus.WithFields(logrus.Fields{
				"contact_uuid": models.ContactID(session.Contact().ID()),
				"group_uuid":   g.UUID,
			}).Warn("unable to find group to remove, skipping")
			continue
		}

		hookEvent := &GroupRemove{
			models.ContactID(session.Contact().ID()),
			group.ID(),
		}

		// add our add event
		session.AddPreCommitEvent(contactGroupsChangedHook, hookEvent)
		session.AddPreCommitEvent(updateCampaignEventsHook, hookEvent)
	}

	return nil
}

// handleContactGroupsAdded is called when a group is added to a contact
func handleContactGroupsAdded(ctx context.Context, tx *sqlx.Tx, rp *redis.Pool, session *models.Session, e flows.Event) error {
	event := e.(*events.ContactGroupsAddedEvent)
	logrus.WithFields(logrus.Fields{
		"contact_uuid": session.ContactUUID(),
		"groups":       event.Groups,
	}).Debug("adding contact to groups")

	// add each of our groups
	for _, g := range event.Groups {
		// look up our group id
		group := session.Org().GroupByUUID(g.UUID)
		if group == nil {
			logrus.WithFields(logrus.Fields{
				"contact_uuid": models.ContactID(session.Contact().ID()),
				"group_uuid":   g.UUID,
			}).Warn("unable to find group to add, skipping")
			continue
		}

		// add our add event
		hookEvent := &GroupAdd{
			models.ContactID(session.Contact().ID()),
			group.ID(),
		}

		session.AddPreCommitEvent(contactGroupsChangedHook, hookEvent)
		session.AddPreCommitEvent(updateCampaignEventsHook, hookEvent)
	}

	return nil
}

// GroupRemove is our struct to track group removals
type GroupRemove struct {
	ContactID models.ContactID `db:"contact_id"`
	GroupID   models.GroupID   `db:"group_id"`
}

const removeContactsFromGroupsSQL = `
DELETE FROM
	contacts_contactgroup_contacts
WHERE 
	id
IN (
	SELECT 
		c.id 
	FROM 
		contacts_contactgroup_contacts c,
		(VALUES(:contact_id, :group_id)) AS g(contact_id, group_id)
	WHERE
		c.contact_id = g.contact_id::int AND c.contactgroup_id = g.group_id::int
);
`

// GroupAdd is our struct to track a final group additions
type GroupAdd struct {
	ContactID models.ContactID `db:"contact_id"`
	GroupID   models.GroupID   `db:"group_id"`
}

const addContactsToGroupsSQL = `
	INSERT INTO 
		contacts_contactgroup_contacts
		(contact_id, contactgroup_id)
	VALUES(:contact_id, :group_id)
	ON CONFLICT
		DO NOTHING
`
