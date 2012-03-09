/*
Package tmpl provides a lazy compilation block-based templating system.

When performing web development, page content and markup are easily visualized
in "blocks." These blocks may include: the header, navigation, main content, or
supporting content. Tmpl provides simple block-based templating system to
accommodate the needs of web development.

Contexts

Contexts are the origin from which all values referenced in a block are
utilized. The main context is passed in when calling Execute to render the
final HTML. From there, sub-contexts are derived by passing into blocks, using
"evoke" or "with".

Values from contexts and sub-contexts are available through the use of
selectors. Selectors always begin with a dot (.), followed by the attribute
name. A single dot selector always references "this" value. Selectors may chain
together to delve deeper into any context, as seen below.

Sub-contexts may always reference their parent context through the use of
dollar signs ($), similar to referencing a parent directory using "..".
Additionally, the top-level context is always available with a leading forward
slash (/).

Given the following main context, represented in JSON format,

	{
		"foo": {
			"bar": {
				"baz": "bif"
			}
		}
	}

if a block is evoked in the following manner,

	{% evoke myBlock .foo.bar.baz %}

then the following selectors will all produce "bif" within "myBlock":

	{% . %}
	{% $.baz %}
	{% $$.bar.baz %}
	{% $$$.foo.bar.baz %}
	{% /.foo.bar.baz %}

Statements

The following outlines the statements available within templates, along with
examples of each.

	{% [/][$[$[...]]].[Selector[.Selector[...]]] %}

Selects a value from a context or sub-context. Selector syntax traverses
contexts in a manner similar to a directory structure:

- Leading forward slash (/) starts selection from the main context

- Dollar sign ($) traverses up the context tree

- Multiple dollar signs ($$) continue to traverse up the context tree

- Dot (.) references "this" value, whether in a main- or sub-context

- Selectors are attribute names which come after dots (.MyValue)

- Multiple selectors are separated by dots (.MyStruct.MySubStruct.MyValue)

	{% block myName %}...{% end block %}

Defines a block with the name, myName. Block definitions must end with an
{% end block %} statement.

	{% evoke myName [context] %}

Substitutes this statement with the contents of the block, myBlock. The
optional context argument pushes a sub-context into the block.

	{% range .Selector [as keyName valueName]}...{% end range %}

Iterates over the value in .Selector. If "as keyName valueName" are present,
the selectors ".keyName" and ".valueName" are available within the range block.
Otherwise, the selectors ".key" and ".val" become available. Similar to the Go
built-in range, "_" is a valid name for either the key or value. Range
definitions must end with an {% end range %} statement. The types which range
will iterate are: map, slice, struct

	{% range call someFunc [as keyName valueName] %}

Similar to ranging over a selector, but first calls the function by the name,
someFunc. All other aspects of operation are identical to the above selector
range.




Tmpl comes with many commands to help with creating dynamic output easily. We
have already seen blocks, evoke, and range. There are also the commands with,
if, and call.

	* Block defines a named block to be inserted by an evoke
	* Evoke calls a named block with an optional context path for it to operate on
	* Range iterates over things like maps/slices/structs setting key/value pairs on the current context path.
	* With temporarily sets the current level of the context path
	* If evaluates the given value and runs one of two outcomes
	* Call runs a user supplied function with the specified arguments

Command Examples

Below you can find a simple example displaying how to use each command.

	// defines a block named foo
	{% block foo %}
		This is a foo block!
	{% end block %}

	// runs the block named foo
	{% evoke foo %}

	// runs the block named foo at the path .bar
	{% evoke foo .bar %}

	// iterates over the value in .iterable printing key/value pairs
	{% range .iterable %}
		{% .key %}: {% .val %}
	{% end range %}

	// iterates over the value in .iterable printing just values
	{% range .iterable as _ v %}
		{% .v %}
	{% end range %}

	// prints the value in .foo
	{% with .foo %}
		{% . %}
	{% end with %}

	// prints if the value in .foo is "truthy"
	{% if .foo %}
		.foo is true!
	{% else %}
		.foo is false!
	{% end if %}

	// calls the function "foo" with a parameter as the value in .bar
	// and displays the result
	{% call foo .bar %}

	// ranges over the value returned by the function call "foo"
	// with a parameter as the value in .bar and displays the key/value pairs
	// of the result.
	{% range call foo .bar as k v %}
		{% .k %}: {% .v %}
	{% end range %}

Modes

Tmpl has two modes, Production and Development, which can be changed at any time
with the CompileMode function. In Development mode, every block and template is 
loaded from disk and compiled in Execute, so that the latest results are always
used. In Production mode, files are only compiled the first time they are needed
and the results are cached for subsequent access.
Example

The template, "base.tmpl", defined as,

	//file: base.tmpl
	<html>
	<head>
		<title>{% .Title %}</title>
		{% evoke meta .Meta %}
	</head>
	<body>
		{% evoke content %}
		<hr>
		{% evoke footer %}
	</body>
	</html>

with the set of blocks,

	//file: footer.block
	{% block footer %}
	We're a footer!
	{% end block %}

	//file: meta.block
	{% block meta %}
		{% range .Javascripts as _ src %}
			<script src="{% .src %}"></script>
		{% end range %}
	{% end block %}

	//file: content.block
	{% block content %}
	Some foofy content with a {% .User.Name %}
	{% end block %}

implementation with tmpl,

	t := tmpl.Parse("base.tmpl")
	t.Blocks("footer.block")
	context := RequestContext(r)
	if err := t.Execute(w, context, "meta.block", "content.block"); err != nil {
		//handle err
	}

and context, represented here as JSON,

	{
		"Meta": {
			"Javascripts": ["one.js", "two.js"]
		},
		"Title": "templates!",
		"User": {
			"Name": "zeebo",
			"Location": "USA"
		}
	}

would lead to the following output (with some whitespace difference):

	<html>
	<head>
		<title>templates!</title>
		<script src="one.js"></script>
		<script src="two.js"></script>
	</head>
	<body>
		Some foofy content with a zeebo
		<hr>
		We're a footer!
	</body>
	</html>

This requests that the "footer.block" file be compiled in for every Execute,
while the "meta.block" and "content.block" files are compiled in for that
specific execute. The block defintions are inserted into the evoke locations
with the current context passed in, or whatever is specified by the evoke.
*/
package tmpl
