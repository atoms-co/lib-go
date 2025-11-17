package log_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"go.cloudkitchens.org/lib/log"
)

func Test_Given_Context_When_TwoNewChildContexts_Then_FieldsAreNotOverwritten(t *testing.T) {
	var ctx1Fields []log.Field
	var ctx2Fields []log.Field
	var parentFields []log.Field
	for i := 0; i < 12; i++ {
		parentFields = append(parentFields, log.Int(fmt.Sprintf("parent-%d", i), i))
		ctx1Fields = append(ctx1Fields, log.Int(fmt.Sprintf("parent-%d", i), i))
		ctx2Fields = append(ctx2Fields, log.Int(fmt.Sprintf("parent-%d", i), i))
	}

	parent := log.NewContext(context.Background(), parentFields...)

	ctx1 := log.NewContext(parent, log.String("child1", "child1-value"))
	ctx2 := log.NewContext(parent, log.String("child2", "child2-value"))

	ctx1Fields = append(ctx1Fields, log.String("child1", "child1-value"))
	ctx2Fields = append(ctx2Fields, log.String("child2", "child2-value"))

	assert.Equal(t, parentFields, log.FromContext(parent))
	assert.Equal(t, ctx1Fields, log.FromContext(ctx1))
	assert.Equal(t, ctx2Fields, log.FromContext(ctx2))
}
