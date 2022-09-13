package bolt

import (
	"encoding/json"

	"github.com/boreq/errors"
	"github.com/planetary-social/scuttlego/service/adapters/bolt/utils"
	"github.com/planetary-social/scuttlego/service/domain/graph"
	"github.com/planetary-social/scuttlego/service/domain/identity"
	"github.com/planetary-social/scuttlego/service/domain/refs"
	"go.etcd.io/bbolt"
)

const socialGraphRepositoryBucket = "graph"

type SocialGraphRepository struct {
	local identity.Public
	hops  graph.Hops
	tx    *bbolt.Tx
}

func NewSocialGraphRepository(tx *bbolt.Tx, local identity.Public, hops graph.Hops) *SocialGraphRepository {
	return &SocialGraphRepository{tx: tx, local: local, hops: hops}
}

func (s *SocialGraphRepository) GetSocialGraph() (*graph.SocialGraph, error) {
	localRef, err := refs.NewIdentityFromPublic(s.local)
	if err != nil {
		return nil, errors.Wrap(err, "could not create a local ref")
	}
	return graph.NewSocialGraph(localRef, s.hops, s)
}

func (s *SocialGraphRepository) Follow(who, contact refs.Identity) error {
	return s.modifyContact(who, contact, func(c *storedContact) {
		c.Following = true
	})
}

func (s *SocialGraphRepository) Unfollow(who, contact refs.Identity) error {
	return s.modifyContact(who, contact, func(c *storedContact) {
		c.Following = false
	})
}

func (s *SocialGraphRepository) Block(who, contact refs.Identity) error {
	return s.modifyContact(who, contact, func(c *storedContact) {
		c.Blocking = true
	})
}

func (s *SocialGraphRepository) Unblock(who refs.Identity, contact refs.Identity) error {
	return s.modifyContact(who, contact, func(c *storedContact) {
		c.Blocking = false
	})
}

func (s *SocialGraphRepository) Remove(who refs.Identity) error {
	return s.deleteFeedBucket(who)
}

func (s *SocialGraphRepository) GetContacts(node refs.Identity) ([]refs.Identity, error) {
	bucket, err := s.getFeedBucket(node)
	if err != nil {
		return nil, errors.Wrap(err, "could not create a bucket")
	}

	if bucket == nil {
		return nil, nil
	}

	var result []refs.Identity

	if err := bucket.ForEach(func(k, v []byte) error {
		contactRef, err := refs.NewIdentity(string(k)) // todo is this certainly a copy or are we reusing the slice illegally
		if err != nil {
			return errors.Wrap(err, "could not create contact ref")
		}

		var c storedContact
		if err := json.Unmarshal(v, &c); err != nil {
			return errors.Wrap(err, "failed to unmarshal the value")
		}

		if c.Following {
			result = append(result, contactRef)
		}

		return nil
	}); err != nil {
		return nil, errors.Wrap(err, "iteration failed")
	}

	return result, nil
}

func (s *SocialGraphRepository) modifyContact(who, contact refs.Identity, f func(c *storedContact)) error {
	bucket, err := s.createFeedBucket(who)
	if err != nil {
		return errors.Wrap(err, "could not create a bucket")
	}

	key := s.key(contact)

	var c storedContact

	value := bucket.Get(key)
	if value != nil {
		if err := json.Unmarshal(value, &c); err != nil {
			return errors.Wrap(err, "failed to unmarshal the existing value")
		}
	}

	f(&c)

	b, err := json.Marshal(c)
	if err != nil {
		return errors.Wrap(err, "could not marshal contact")
	}

	return bucket.Put(key, b)

}

func (s *SocialGraphRepository) createFeedBucket(ref refs.Identity) (*bbolt.Bucket, error) {
	return utils.CreateBucket(s.tx, s.pathFunc(ref))
}

func (s *SocialGraphRepository) getFeedBucket(ref refs.Identity) (*bbolt.Bucket, error) {
	return utils.GetBucket(s.tx, s.pathFunc(ref))
}

func (s *SocialGraphRepository) deleteFeedBucket(ref refs.Identity) error {
	return utils.DeleteBucket(
		s.tx,
		[]utils.BucketName{
			utils.BucketName(socialGraphRepositoryBucket),
		},
		utils.BucketName(ref.String()),
	)
}

func (s *SocialGraphRepository) pathFunc(who refs.Identity) []utils.BucketName {
	return []utils.BucketName{
		utils.BucketName(socialGraphRepositoryBucket),
		utils.BucketName(who.String()),
	}
}

func (s *SocialGraphRepository) key(target refs.Identity) utils.BucketName {
	return []byte(target.String())
}

type storedContact struct {
	Following bool `json:"following"`
	Blocking  bool `json:"blocking"`
}
