package messages

import (
	"encoding/json"

	"github.com/boreq/errors"
	"github.com/planetary-social/scuttlego/service/domain/refs"
	"github.com/planetary-social/scuttlego/service/domain/rooms"
	"github.com/planetary-social/scuttlego/service/domain/transport/rpc"
)

var (
	RoomListAliasesProcedure = rpc.MustNewProcedure(
		rpc.MustNewProcedureName([]string{"room", "listAliases"}),
		rpc.ProcedureTypeAsync,
	)
)

func NewRoomListAliases(arguments RoomListAliasesArguments) (*rpc.Request, error) {
	j, err := arguments.MarshalJSON()
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal arguments")
	}

	return rpc.NewRequest(
		RoomListAliasesProcedure.Name(),
		RoomListAliasesProcedure.Typ(),
		j,
	)
}

type RoomListAliasesArguments struct {
	identity refs.Identity
}

func NewRoomListAliasesArguments(
	identity refs.Identity,
) (RoomListAliasesArguments, error) {
	if identity.IsZero() {
		return RoomListAliasesArguments{}, errors.New("zero value of identity")
	}

	return RoomListAliasesArguments{
		identity: identity,
	}, nil
}

func (i RoomListAliasesArguments) MarshalJSON() ([]byte, error) {
	return json.Marshal([]string{
		i.identity.String(),
	})
}

type RoomListAliasesResponse struct {
	aliases []rooms.Alias
}

func NewRoomsListAliasesResponse(b []byte) (RoomListAliasesResponse, error) {
	var aliasesAsStrings []string
	if err := json.Unmarshal(b, &aliasesAsStrings); err != nil {
		return RoomListAliasesResponse{}, errors.Wrap(err, "json unmarshal failed")
	}

	var aliases []rooms.Alias
	for _, aliasString := range aliasesAsStrings {
		alias, err := rooms.NewAlias(aliasString)
		if err != nil {
			return RoomListAliasesResponse{}, errors.Wrap(err, "error creating an alias")
		}
		aliases = append(aliases, alias)
	}

	return RoomListAliasesResponse{aliases: aliases}, nil
}

func (r RoomListAliasesResponse) Aliases() []rooms.Alias {
	return r.aliases
}
