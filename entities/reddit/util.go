package reddit

import (
	"strconv"
	"strings"
	"time"
)

type jsonTimestamp time.Time

func (t *jsonTimestamp) UnmarshalJSON(s []byte) (err error) {
	r := strings.TrimSuffix(string(s), ".0")
	q, err := strconv.ParseInt(r, 10, 64)
	if err != nil {
		return err
	}
	*(*time.Time)(t) = time.Unix(q, 0)
	return nil
}

func (t jsonTimestamp) String() string {
	return time.Time(t).String()
}
