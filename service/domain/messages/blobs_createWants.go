package messages

import (
	"fmt"
	"strings"

	"github.com/boreq/errors"
	jsoniter "github.com/json-iterator/go"
	"github.com/planetary-social/scuttlego/service/domain/blobs"
	"github.com/planetary-social/scuttlego/service/domain/refs"
	"github.com/planetary-social/scuttlego/service/domain/transport/rpc"
)

var (
	BlobsCreateWantsProcedure = rpc.MustNewProcedure(
		rpc.MustNewProcedureName([]string{"blobs", "createWants"}),
		rpc.ProcedureTypeSource,
	)
)

func NewBlobsCreateWants() (*rpc.Request, error) {
	return rpc.NewRequest(
		BlobsCreateWantsProcedure.Name(),
		BlobsCreateWantsProcedure.Typ(),
		[]byte("[]"),
	)
}

type BlobsCreateWantsResponse struct {
	v []BlobWithSizeOrWantDistance
}

func NewBlobsCreateWantsResponseFromBytes(b []byte) (BlobsCreateWantsResponse, error) {
	m := make(blobsCreateWantsResponseTransport)

	if err := jsoniter.Unmarshal(b, &m); err != nil {
		return BlobsCreateWantsResponse{}, errors.Wrap(err, "json unmarshal failed")
	}

	var elements []BlobWithSizeOrWantDistance

	for blobRefString, sizeOrWantDistance := range m {
		id, err := refs.NewBlob(blobRefString)
		if err != nil {
			return BlobsCreateWantsResponse{}, errors.New("invalid blob ref")
		}

		sizeOrWantDistance, err := blobs.NewSizeOrWantDistance(sizeOrWantDistance)
		if err != nil {
			return BlobsCreateWantsResponse{}, errors.New("invalid size or want distance")
		}

		element, err := NewBlobWithSizeOrWantDistance(id, sizeOrWantDistance)
		if err != nil {
			return BlobsCreateWantsResponse{}, errors.New("could not create an element")
		}

		elements = append(elements, element)
	}

	return BlobsCreateWantsResponse{
		v: elements,
	}, nil
}

func NewBlobsCreateWantsResponse(id refs.Blob, sizeOrWantDistance blobs.SizeOrWantDistance) (BlobsCreateWantsResponse, error) {
	element, err := NewBlobWithSizeOrWantDistance(id, sizeOrWantDistance)
	if err != nil {
		return BlobsCreateWantsResponse{}, errors.New("could not create an element")
	}

	return BlobsCreateWantsResponse{
		v: []BlobWithSizeOrWantDistance{
			element,
		},
	}, nil
}

func (resp BlobsCreateWantsResponse) List() []BlobWithSizeOrWantDistance {
	tmp := make([]BlobWithSizeOrWantDistance, len(resp.v))
	copy(tmp, resp.v)
	return tmp
}

func (resp BlobsCreateWantsResponse) MarshalJSON() ([]byte, error) {
	m := make(blobsCreateWantsResponseTransport)

	for _, element := range resp.v {
		if s, ok := element.SizeOrWantDistance().Size(); ok {
			m[element.Id().String()] = s.InBytes()
			continue
		}

		if d, ok := element.SizeOrWantDistance().WantDistance(); ok {
			m[element.Id().String()] = -int64(d.Int())
			continue
		}

		panic("not all cases are covered")
	}

	return jsoniter.Marshal(m)
}

type blobsCreateWantsResponseTransport map[string]int64

type BlobWithSizeOrWantDistance struct {
	id                 refs.Blob
	sizeOrWantDistance blobs.SizeOrWantDistance
}

func NewBlobWithSizeOrWantDistance(id refs.Blob, sizeOrWantDistance blobs.SizeOrWantDistance) (BlobWithSizeOrWantDistance, error) {
	if id.IsZero() {
		return BlobWithSizeOrWantDistance{}, errors.New("zero value of id")
	}

	if sizeOrWantDistance.IsZero() {
		return BlobWithSizeOrWantDistance{}, errors.New("zero value of size or want distance")
	}

	return BlobWithSizeOrWantDistance{
		id:                 id,
		sizeOrWantDistance: sizeOrWantDistance,
	}, nil
}

func NewBlobWithWantDistance(id refs.Blob, wantDistance blobs.WantDistance) (BlobWithSizeOrWantDistance, error) {
	v, err := blobs.NewSizeOrWantDistanceContainingWantDistance(wantDistance)
	if err != nil {
		return BlobWithSizeOrWantDistance{}, errors.Wrap(err, "invalid want distance")
	}

	return NewBlobWithSizeOrWantDistance(id, v)
}

func MustNewBlobWithWantDistance(id refs.Blob, wantDistance blobs.WantDistance) BlobWithSizeOrWantDistance {
	v, err := NewBlobWithWantDistance(id, wantDistance)
	if err != nil {
		panic(err)
	}
	return v
}

func NewBlobWithSize(id refs.Blob, size blobs.Size) (BlobWithSizeOrWantDistance, error) {
	v, err := blobs.NewSizeOrWantDistanceContainingSize(size)
	if err != nil {
		return BlobWithSizeOrWantDistance{}, errors.Wrap(err, "invalid size")
	}

	return NewBlobWithSizeOrWantDistance(id, v)
}

func MustNewBlobWithSize(id refs.Blob, size blobs.Size) BlobWithSizeOrWantDistance {
	v, err := NewBlobWithSize(id, size)
	if err != nil {
		panic(err)
	}
	return v
}

func (b BlobWithSizeOrWantDistance) Id() refs.Blob {
	return b.id
}

func (b BlobWithSizeOrWantDistance) SizeOrWantDistance() blobs.SizeOrWantDistance {
	return b.sizeOrWantDistance
}

func (b BlobWithSizeOrWantDistance) String() string {
	var s []string
	s = append(s, fmt.Sprintf("id=%s", b.id.String()))
	if distance, ok := b.sizeOrWantDistance.WantDistance(); ok {
		s = append(s, fmt.Sprintf("distance=%d", distance.Int()))
	}
	if size, ok := b.sizeOrWantDistance.Size(); ok {
		s = append(s, fmt.Sprintf("size=%d", size.InBytes()))
	}
	return fmt.Sprintf("<%s>", strings.Join(s, " "))
}
