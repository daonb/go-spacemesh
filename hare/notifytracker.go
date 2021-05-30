package hare

import (
	"encoding/binary"
	"hash/fnv"
)

// notifyTracker tracks notify messages.
// It also provides the number of notifications tracked for a given set.
type notifyTracker struct {
	notifies     map[string]struct{} // tracks PubKey->Notification
	tracker      *RefCountTracker    // tracks ref count to each seen set
	certificates map[uint32]struct{} // tracks Set->certificate
	//also store the notify messages themselves, for the certificate
	notificationMessages map[uint32][]*Message // tracks Set->Messages
}

func newNotifyTracker(expectedSize int) *notifyTracker {
	nt := &notifyTracker{}
	nt.notifies = make(map[string]struct{}, expectedSize)
	nt.tracker = NewRefCountTracker()
	nt.certificates = make(map[uint32]struct{}, expectedSize)
	nt.notificationMessages = make(map[uint32][]*Message)

	return nt
}

// OnNotify tracks the provided notification message.
// Returns true if the InnerMsg didn't affect the state, false otherwise
func (nt *notifyTracker) OnNotify(msg *Msg) bool {
	pub := msg.PubKey
	eligibilityCount := uint32(msg.InnerMsg.EligibilityCount)
	if _, exist := nt.notifies[pub.String()]; exist { // already seenSenders
		return true // ignored
	}

	// keep msg for pub
	nt.notifies[pub.String()] = struct{}{}

	// track that set
	s := NewSet(msg.InnerMsg.Values)
	nt.onCertificate(msg.InnerMsg.Cert.AggMsgs.Messages[0].InnerMsg.K, s)
	setID := s.ID()
	nt.tracker.Track(setID, eligibilityCount)

	//store the message to create a certificate for the certification messages
	if _, exists := nt.notificationMessages[setID]; !exists {
		nt.notificationMessages[setID] = make([]*Message, 0)
	}
	nt.notificationMessages[setID] = append(nt.notificationMessages[setID], msg.Message)
	// nt.tracker.Track(s.ID(), eligibilityCount)

	return false
}

// NotificationsCount returns the number of notifications tracked for the provided set
func (nt *notifyTracker) NotificationsCount(s *Set) int {
	return int(nt.tracker.CountStatus(s.ID()))
}

// calculates a unique id for the provided k and set.
func calcID(k int32, set *Set) uint32 {
	hash := fnv.New32()

	// write K
	buff := make([]byte, 4)
	binary.LittleEndian.PutUint32(buff, uint32(k)) // K>=0 because this not pre-round
	hash.Write(buff)

	// TODO: is this hash enough for this usage?
	// write set ObjectID
	buff = make([]byte, 4)
	binary.LittleEndian.PutUint32(buff, uint32(set.ID()))
	hash.Write(buff)

	return hash.Sum32()
}

// tracks certificates
func (nt *notifyTracker) onCertificate(k int32, set *Set) {
	nt.certificates[calcID(k, set)] = struct{}{}
}

// HasCertificate returns true if a certificate exist for the provided set in the provided round, false otherwise.
func (nt *notifyTracker) HasCertificate(k int32, set *Set) bool {
	_, exist := nt.certificates[calcID(k, set)]
	return exist
}

// BuildCertificate builds a certificate of all the notify messages to send as part of the termination certification
func (nt *notifyTracker) BuildCertificate(s *Set) *Certificate {
	// assume for now that you'll always be able to make a valid certificate
	c := &Certificate{}
	c.Values = s.ToSlice()
	c.AggMsgs = &aggregatedMessages{}
	c.AggMsgs.Messages = nt.notificationMessages[s.ID()]

	//figure out if any further optimizations have to occur
	for _, commit := range c.AggMsgs.Messages {
		commit.InnerMsg.Values = nil
	}

	return c
}
