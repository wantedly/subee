package generator

import (
	"bytes"
	"context"
	"fmt"
	"go/types"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/iancoleman/strcase"
	"github.com/logrusorgru/aurora"
	"github.com/pkg/errors"
	"golang.org/x/tools/go/packages"
	"golang.org/x/tools/imports"
)

type SubscriberGenerator interface {
	Generate(context.Context, *SubscriberParams) error
}

func NewSubscriberGenerator(
	pkgCfg *packages.Config,
	outW io.Writer,
) SubscriberGenerator {
	return &subscriberGeneratorImpl{
		pkgCfg: pkgCfg,
		outW:   outW,
	}
}

type subscriberGeneratorImpl struct {
	pkgCfg *packages.Config
	outW   io.Writer
}

func (g *subscriberGeneratorImpl) Generate(ctx context.Context, params *SubscriberParams) error {
	if err := params.Validate(); err != nil {
		return errors.WithStack(err)
	}
	if g.pkgCfg.Mode < packages.LoadTypes {
		return errors.Errorf("invalid packages.LoadMode: %v", g.pkgCfg.Mode)
	}

	if params.IsWithAdapter() {
		pkgs, err := packages.Load(g.pkgCfg, params.Package.Path)
		if err != nil {
			return errors.WithStack(err)
		}
		var obj types.Object
		for _, pkg := range pkgs {
			obj = pkg.Types.Scope().Lookup(params.Message)
			if obj == nil {
				continue
			}
			if _, ok := obj.Type().Underlying().(*types.Struct); !ok {
				continue
			}
			break
		}
		if obj == nil {
			return errors.Errorf("Message type %q not found in the given package path %q", params.Message, params.Package.Path)
		}
		params.Package.Path = obj.Pkg().Path()
		params.Package.Name = obj.Pkg().Name()
		if filepath.Base(params.Package.Path) == params.Package.Name {
			params.Imports = append(params.Imports, Package{Path: params.Package.Path})
		} else {
			params.Imports = append(params.Imports, params.Package)
		}
		if params.Encoding == MessageEncodingProtobuf {
			params.Imports = append(params.Imports, Package{Path: "github.com/golang/protobuf/proto"})
		}
	}

	if params.Batch {
		err := g.writeFile(filepath.Join("pkg", "consumer", strcase.ToSnake(params.Name)+"_batch_consumer.go"), params, batchConsumerTmpl)
		if err != nil {
			return errors.WithStack(err)
		}

		if params.IsWithAdapter() {
			err = g.writeFile(filepath.Join("pkg", "consumer", strcase.ToSnake(params.Name)+"_batch_consumer_adapter.go"), params, batchAdapterTmpl)
			if err != nil {
				return errors.WithStack(err)
			}
		}
	} else {
		err := g.writeFile(filepath.Join("pkg", "consumer", strcase.ToSnake(params.Name)+"_consumer.go"), params, consumerTmpl)
		if err != nil {
			return errors.WithStack(err)
		}

		if params.IsWithAdapter() {
			err = g.writeFile(filepath.Join("pkg", "consumer", strcase.ToSnake(params.Name)+"_consumer_adapter.go"), params, adapterTmpl)
			if err != nil {
				return errors.WithStack(err)
			}
		}
	}

	err := g.writeFile(filepath.Join("cmd", strcase.ToKebab(params.Name)+"-subscriber", "run.go"), params, runTmpl)
	if err != nil {
		return errors.WithStack(err)
	}

	err = g.writeFile(filepath.Join("cmd", strcase.ToKebab(params.Name)+"-subscriber", "main.go"), params, mainTmpl)
	if err != nil {
		return errors.WithStack(err)
	}

	{
		strong := func(arg interface{}) aurora.Value { return aurora.BrightWhite(arg).Bold() }
		bullet := aurora.Yellow("▸")
		fmt.Fprintln(g.outW)
		fmt.Fprintln(g.outW, "Scaffold", aurora.BrightWhite(strcase.ToKebab(params.Name)+"-subscriber"), "successfully 🎉")
		consumer := "./pkg/consumer/" + strcase.ToSnake(params.Name) + "_consumer.go"
		if params.Batch {
			consumer = strings.Replace(consumer, "_consumer.go", "_batch_consumer.go", 1)
		}
		fmt.Fprintln(g.outW, bullet, "At first, you should implement", strong("createSubscriber()"), "in", strong("./cmd/"+strcase.ToKebab(params.Name)+"-subscriber/run.go"))
		fmt.Fprintln(g.outW, bullet, "You can implement a messages handler in", strong(consumer))
		fmt.Fprintln(g.outW, bullet, "You can run subscribers with", strong("subee start"), "command")
	}

	return nil
}

func mustCreateTemplate(name, text string) *template.Template {
	return template.Must(template.New(name).Funcs(tmplFuncs).Parse(text))
}

func (g *subscriberGeneratorImpl) writeFile(path string, params interface{}, tmpl *template.Template) error {
	buf := new(bytes.Buffer)
	err := tmpl.Execute(buf, params)
	if err != nil {
		return errors.WithStack(err)
	}
	abspath := filepath.Join(g.pkgCfg.Dir, path)
	data, err := imports.Process(abspath, buf.Bytes(), nil)
	if err != nil {
		return errors.WithStack(err)
	}
	if _, err := os.Stat(filepath.Dir(abspath)); err != nil {
		err = os.MkdirAll(filepath.Dir(abspath), 0755)
		if err != nil {
			return errors.WithStack(err)
		}
	}
	err = ioutil.WriteFile(abspath, data, 0644)
	if err != nil {
		return errors.WithStack(err)
	}

	fmt.Fprintf(g.outW, "%4s %s\n", aurora.Green("✔"), path)
	return nil
}

var (
	tmplFuncs = template.FuncMap{
		"ToCamel":      strcase.ToCamel,
		"ToLowerCamel": strcase.ToLowerCamel,
	}

	mainTmpl = mustCreateTemplate("cmd/{{.Name}}/main.go",
		`// Code generated by github.com/wantedly/subee/cmd/subee. DO NOT EDIT.

package main

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}`)

	runTmpl = mustCreateTemplate("cmd/{{.Name}}/run.go",
		`package main

func run() error {
	ctx := context.Background()

	subscriber, err := createSubscriber(ctx)
	if err != nil {
		return errors.WithStack(err)
	}

	{{if .Batch}}
	engine := subee.NewBatch(
	{{- else}}
	engine := subee.New(
	{{- end}}
		subscriber,
		{{- if and .Batch .IsWithAdapter}}
		consumer.New{{.Name|ToCamel}}BatchConsumerAdapter(
			consumer.New{{.Name|ToCamel}}BatchConsumer(),
		),
		{{- else if .Batch}}
		consumer.New{{.Name|ToCamel}}BatchConsumer(),
		{{- else if .IsWithAdapter}}
		consumer.New{{.Name|ToCamel}}ConsumerAdapter(
			consumer.New{{.Name|ToCamel}}Consumer(),
		),
		{{- else}}
		consumer.New{{.Name|ToCamel}}Consumer(),
		{{- end}}
	)

	return engine.Start(ctx)
}

func createSubscriber(ctx context.Context) (subee.Subscriber, error) {
	// TODO: not yet implemented
	return nil, errors.New("createSubscriber() has not been implemented yet")
}`)

	consumerTmpl = mustCreateTemplate("pkg/consumer/{{.Name}}_consumer.go",
		`package consumer

import (
{{- range .Imports}}
	{{- if .Name}}
		{{.Name}} "{{.Path}}"
	{{- else}}
		"{{.Path}}"
	{{- end}}
{{- end}}
)

{{- if .IsWithAdapter}}
// {{.Name|ToCamel}}Consumer is a consumer interface for {{.Package.Name}}.{{.Message}}.
type {{.Name|ToCamel}}Consumer interface {
	Consume(context.Context, *{{.Package.Name}}.{{.Message}}) error
}
{{- else}}
// {{.Name|ToCamel}}Consumer is a consumer interface.
type {{.Name|ToCamel}}Consumer subee.Consumer
{{- end}}

// New{{.Name|ToCamel}}Consumer creates a new consumer instance.
func New{{.Name|ToCamel}}Consumer() {{.Name|ToCamel}}Consumer {
	return &{{.Name|ToLowerCamel}}ConsumerImpl{}
}

type {{.Name|ToLowerCamel}}ConsumerImpl struct {}

{{- if .IsWithAdapter}}
func (c *{{.Name|ToLowerCamel}}ConsumerImpl) Consume(ctx context.Context, msg *{{.Package.Name}}.{{.Message}}) error {
{{- else}}
func (c *{{.Name|ToLowerCamel}}ConsumerImpl) Consume(ctx context.Context, msg subee.Message) error {
{{- end}}
	return errors.New("Consume() has not been implemented yet")
}`)

	batchConsumerTmpl = mustCreateTemplate("pkg/consumer/{{.Name}}_batch_consumer.go",
		`package consumer

import (
{{- range .Imports}}
	{{- if .Name}}
		{{.Name}} "{{.Path}}"
	{{- else}}
		"{{.Path}}"
	{{- end}}
{{- end}}
)

{{- if .IsWithAdapter}}
// {{.Name|ToCamel}}BatchConsumer is a batch consumer interface for {{.Package.Name}}.{{.Message}}.
type {{.Name|ToCamel}}BatchConsumer interface {
	BatchConsume(context.Context, []*{{.Package.Name}}.{{.Message}}) error
}
{{- else}}
// {{.Name|ToCamel}}BatchConsumer is a consumer interface.
type {{.Name|ToCamel}}BatchConsumer subee.BatchConsumer
{{- end}}

// New{{.Name|ToCamel}}BatchConsumer creates a new consumer instance.
func New{{.Name|ToCamel}}BatchConsumer() {{.Name|ToCamel}}BatchConsumer {
	return &{{.Name|ToLowerCamel}}BatchConsumerImpl{}
}

type {{.Name|ToLowerCamel}}BatchConsumerImpl struct {}

{{- if .IsWithAdapter}}
func (c *{{.Name|ToLowerCamel}}BatchConsumerImpl) BatchConsume(ctx context.Context, msgs []*{{.Package.Name}}.{{.Message}}) error {
{{- else}}
func (c *{{.Name|ToLowerCamel}}BatchConsumerImpl) BatchConsume(ctx context.Context, msgs []subee.Message) error {
{{- end}}
	return errors.New("BatchConsume() has not been implemented yet")
}`)

	adapterTmpl = mustCreateTemplate("pkg/consumer/{{.Name}}_consumer_adapter.go",
		`// Code generated by github.com/wantedly/subee/cmd/subee. DO NOT EDIT.

package consumer

import (
{{- range .Imports}}
	{{- if .Name}}
		{{.Name}} "{{.Path}}"
	{{- else}}
		"{{.Path}}"
	{{- end}}
{{- end}}
)

// New{{.Name|ToCamel}}ConsumerAdapter created a consumer-adapter instance that converts incoming messages into {{.Package.Name}}.{{.Message}}.
func New{{.Name|ToCamel}}ConsumerAdapter(consumer {{.Name|ToCamel}}Consumer) subee.Consumer {
	return &{{.Name|ToLowerCamel}}ConsumerAdapterImpl{consumer: consumer}
}

type {{.Name|ToLowerCamel}}ConsumerAdapterImpl struct {
	consumer {{.Name|ToCamel}}Consumer
}

func (a *{{.Name|ToLowerCamel}}ConsumerAdapterImpl) Consume(ctx context.Context, m subee.Message) error {
	var err error
	obj := new({{.Package.Name}}.{{.Message}})
	{{- if .IsJSON}}
	err = json.Unmarshal(m.Data(), obj)
	{{- else if .IsProtobuf}}
	err = proto.Unmarshal(m.Data(), obj)
	{{- end}}
	if err != nil {
		return errors.WithStack(err)
	}
	err = a.consumer.Consume(ctx, obj)
	if err != nil {
		return errors.WithStack(err)
	}
	return nil
}`)

	batchAdapterTmpl = mustCreateTemplate("pkg/consumer/{{.Name}}_batch_consumer_adapter.go",
		`// Code generated by github.com/wantedly/subee/cmd/subee. DO NOT EDIT.

package consumer

import (
{{- range .Imports}}
	{{- if .Name}}
		{{.Name}} "{{.Path}}"
	{{- else}}
		"{{.Path}}"
	{{- end}}
{{- end}}
)

// New{{.Name|ToCamel}}BatchConsumerAdapter created a consumer-adapter instance that converts incoming messages into {{.Package.Name}}.{{.Message}}.
func New{{.Name|ToCamel}}BatchConsumerAdapter(consumer {{.Name|ToCamel}}BatchConsumer) subee.BatchConsumer {
	return &{{.Name|ToLowerCamel}}BatchConsumerAdapterImpl{consumer: consumer}
}

type {{.Name|ToLowerCamel}}BatchConsumerAdapterImpl struct {
	consumer {{.Name|ToCamel}}BatchConsumer
}

func (a *{{.Name|ToLowerCamel}}BatchConsumerAdapterImpl) BatchConsume(ctx context.Context, ms []subee.Message) error {
	var err error
	objs := make([]*{{.Package.Name}}.{{.Message}}, len(ms))
	for i, m := range ms {
		obj := new({{.Package.Name}}.{{.Message}})
		{{- if .IsJSON}}
		err = json.Unmarshal(m.Data(), obj)
		{{- else if .IsProtobuf}}
		err = proto.Unmarshal(m.Data(), obj)
		{{- end}}
		if err != nil {
			return errors.WithStack(err)
		}
		objs[i] = obj
	}
	err = a.consumer.BatchConsume(ctx, objs)
	if err != nil {
		return errors.WithStack(err)
	}
	return nil
}`)
)

type SubscriberParams struct {
	Package  Package
	Message  string
	Encoding MessageEncoding
	Name     string
	Batch    bool
	Imports  []Package
}

func (p *SubscriberParams) Validate() error {
	switch {
	case p.Package.Path != "" && p.Message == "":
		return fmt.Errorf("message is required")
	case p.Package.Path == "" && p.Message != "":
		return fmt.Errorf("package is required")
	case p.Name == "":
		return fmt.Errorf("name is required")
	}
	return nil
}

func (p *SubscriberParams) IsWithAdapter() bool {
	return p.Package.Path != "" && p.Message != ""
}

func (p *SubscriberParams) IsJSON() bool {
	return p.Encoding == MessageEncodingJSON
}

func (p *SubscriberParams) IsProtobuf() bool {
	return p.Encoding == MessageEncodingProtobuf
}

type MessageEncoding int

const (
	MessageEncodingUnknown MessageEncoding = iota
	MessageEncodingJSON
	MessageEncodingProtobuf
)

var (
	nameByMessageEncoding = map[MessageEncoding]string{}
	messageTypeByName     = map[string]MessageEncoding{}
)

func init() {
	nameByMessageEncoding = map[MessageEncoding]string{
		MessageEncodingJSON:     "json",
		MessageEncodingProtobuf: "protobuf",
	}
	for e, s := range nameByMessageEncoding {
		messageTypeByName[s] = e
	}
}

func (me MessageEncoding) String() string {
	if s, ok := nameByMessageEncoding[me]; ok {
		return s
	}
	return "unknown"
}

func (me *MessageEncoding) Set(in string) error {
	if v, ok := messageTypeByName[in]; ok {
		*me = v
		return nil
	}
	return fmt.Errorf("unknown message type %q", in)
}

func (MessageEncoding) Type() string {
	return "MessageEncoding"
}

type Package struct {
	Name string
	Path string
}
