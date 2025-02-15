package queries_test

import (
	"testing"

	"github.com/boreq/errors"
	"github.com/planetary-social/scuttlego/internal"
	"github.com/planetary-social/scuttlego/internal/fixtures"
	"github.com/planetary-social/scuttlego/service/app/queries"
	"github.com/planetary-social/scuttlego/service/di"
	"github.com/planetary-social/scuttlego/service/domain/blobs"
	"github.com/stretchr/testify/require"
)

func TestGetBlob(t *testing.T) {
	id := fixtures.SomeRefBlob()
	data := fixtures.SomeBytes()
	correctSize := blobs.MustNewSize(int64(len(data)))
	incorrectSize := blobs.MustNewSize(correctSize.InBytes() + int64(fixtures.SomePositiveInt()))

	testCases := []struct {
		Name          string
		Query         queries.GetBlob
		ExpectedError error
	}{
		{
			Name: "only_id",
			Query: queries.GetBlob{
				Id:   id,
				Size: nil,
				Max:  nil,
			},
			ExpectedError: nil,
		},
		{
			Name: "id_and_correct_size",
			Query: queries.GetBlob{
				Id:   id,
				Size: internal.Ptr(correctSize),
				Max:  nil,
			},
			ExpectedError: nil,
		},
		{
			Name: "id_and_incorrect_size",
			Query: queries.GetBlob{
				Id:   id,
				Size: internal.Ptr(incorrectSize),
				Max:  nil,
			},
			ExpectedError: errors.New("blob size doesn't match the provided size"),
		},
		{
			Name: "id_and_max_above_size",
			Query: queries.GetBlob{
				Id:   id,
				Size: nil,
				Max:  internal.Ptr(blobs.MustNewSize(correctSize.InBytes() + 1)),
			},
			ExpectedError: nil,
		},
		{
			Name: "id_and_max_equal_to_size",
			Query: queries.GetBlob{
				Id:   id,
				Size: nil,
				Max:  internal.Ptr(correctSize),
			},
			ExpectedError: nil,
		},
		{
			Name: "id_and_max_below_size",
			Query: queries.GetBlob{
				Id:   id,
				Size: nil,
				Max:  internal.Ptr(blobs.MustNewSize(correctSize.InBytes() - 1)),
			},
			ExpectedError: errors.New("blob is larger than the provided max size"),
		},
		{
			Name: "size_wins_over_max",
			Query: queries.GetBlob{
				Id:   id,
				Size: internal.Ptr(blobs.MustNewSize(1)),
				Max:  internal.Ptr(blobs.MustNewSize(1)),
			},
			ExpectedError: errors.New("blob size doesn't match the provided size"),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			q, err := di.BuildTestQueries(t)
			require.NoError(t, err)

			q.BlobStorage.MockBlob(id, data)

			t.Log(testCase.Query.Size)
			rc, err := q.Queries.GetBlob.Handle(testCase.Query)
			if testCase.ExpectedError != nil {
				require.EqualError(t, err, testCase.ExpectedError.Error())
				require.Nil(t, rc)
			} else {
				require.NoError(t, err)
				require.NotNil(t, rc)
			}
		})
	}
}
