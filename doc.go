/*
Package tmpl provides a lazy compilation block based templating system.

Tmpl is a simple block based templating system. What does block based mean? When
doing web design typically one thinks of "blocks" of content, for example in a
navigation bar, the main content section of the page, or the header section of
the page. Blocks are "evoked" by a main template, and defined in supporting files.

Example

One could have the template defined in "base.tmpl"

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

One could then load an execute this template by

	t := tmpl.Parse("base.tmpl")
	t.Blocks("footer.block")
	context := RequestContext(r)
	if err := t.Execute(w, context, "meta.block", "content.block"); err != nil {
		//handle err
	}

This requests that the "footer.block" file be compiled in for every Execute,
while the "meta.block" and "content.block" files are compiled in for that
specific execute. The block defintions are inserted into the evoke locations
with the current context passed in, or whatever is specified by the evoke. Thus
for some context represented in json as

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

we would expect the output (with some whitespace difference)

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

Discussion

Let's break this example down. First we'll start with the idea of a "context".
A template is passed in a value called the "context" in the Execute call. Values
from the context can be accessed by using "selectors". For example in the 
"base.tmpl" file, we use a relative selector ".Title" to print out the title,
and ".Meta" to specify a spot in the context to anchor our block evocation. So
when the "meta.block" template is called, everything it can select is below
the ".Meta" level in the context. But what if we wanted something above the
passed in context? Do we have to pass in the highest level and have it select
down for everything? Fortunately, no, there are two ways to break out and look
"above" what was passed in. Absolute selectors, which are prefixed with a "/"
and "popping" selectors which are prefixed with a number of "$". This sounds
complicated, but assuming our context looks like

	{
		"foo": {
			"bar": {
				"baz": "baz"
			}
		}
	}

And our context is "rooted" at ".foo.bar.baz", we can access the string "baz"
in these ways:

	{% . %}
	{% $.baz %}
	{% $$.bar.baz %}
	{% $$$.foo.bar.baz %}
	{% /.foo.bar.baz %}

The last of which is an absolute selector. So if we wanted to reference the title
of the page inside of the meta block, we could use {% $.Title %}. A good metaphor
to use for contexts is to think of them like a directory path, with a current
working directory.

Commands

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
*/
package tmpl
