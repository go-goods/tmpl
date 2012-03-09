/*
Package tmpl provides a lazy compilation block-based templating system.

When performing web development, page content and markup are easily visualized
in "blocks." These blocks may include: the header, navigation, main content, or
supporting content. Tmpl provides simple block-based templating system to
accommodate the needs of web development.

Statements

Template statements are surrounded by the tokens '{%' and '%}' and contain text
to specify the action the template system should take. All actions begin with
a keyword like "block" or "evoke", except in the case of printing a value from
a context, where just the selector is specified.

Contexts

Contexts are the origin for all of the values a template has access to. The main
context is passed in when calling Execute to render the final output. From there, 
sub-contexts are derived by passing into blocks, using "evoke" or "with".

Statement - Selector

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

and assuming the statements are being executed in a sub-context rooted at the
"baz" element, then the following selectors will all produce "bif":

	{% . %}
	{% $.baz %}
	{% $$.bar.baz %}
	{% $$$.foo.bar.baz %}
	{% /.foo.bar.baz %}

Statement - Call

Call runs a function that is attached to the template before it is Executed with
the supplied arguments. For example if we had a function named "foo" that was
attached to the template, the action

	{% call foo .Bar .Baz %}

corresponds to the function call

	foo(.Bar, .Baz)

with the selectors .Bar and .Baz evaluated. See Template.Call for details on
how to attach a function.

	{% call name [args...] %}

	{% call titleCase .Title %}
	{% call add .FirstNumber .SecondNumber %}
	{% call not .Value %}
	{% call equal .FirstName .OtherUser %}

Statement - Block

Defines a block with the name, myName. Block definitions must end with an
{% end block %} statement.

	{% block myName %}...{% end block %}

	{% block greeting %}Hello!{% end block %}
	{% block fullName %}{% .FirstName %} {% .LastName %}{% end block %}

Statement - Evoke

Substitutes this statement with the contents of the block, myBlock. The
optional context argument pushes a sub-context into the block.

	{% evoke myName [context] %}

	{% evoke greeting %}
	{% evoke fullName .LoggedInUser %}

Statement - Range

Iterates over the given value. A value can be either the result of a .Selector
or the result of a call statement. If "as keyName valueName" are present,
the selectors ".keyName" and ".valueName" are available within the range block.
Otherwise, the selectors ".key" and ".val" become available. Similar to the Go
built-in range, "_" is a valid name for either the key or value. Range
definitions must end with an {% end range %} statement. The types which range
will iterate are: map, slice, struct

	{% range value [as keyName valueName]}...{% end range %}

	{% range .LoggedInUser.Friends %}
		{% evoke fullName .val %}
	{% end range %}

	{% range .LoggedInUser.Friends as _ friend %}
		{% evoke fullName .friend %}
	{% end range %}

	{% range call someFunc [as keyName valueName] %}...{% end range %}

	{% range call loggedInUsers %}
		{% evoke fullName .val %}
	{% end range %}

	{% range call loggedInUsers as _ user %}
		{% evoke fullName .user %}
	{% end range %}

Statement - If

Evaluates the specified value which may be either a .Selector or the result of
a call statement. If the value is "truthy" it executes the postive template,
otherwise, it executes the negative template if given.

	{% if value %}...[{% else %}...]{% end if %}

	{% if .LoggedIn %}
		Positive: {% evoke fullName .LoggedInUser %} is logged in!
	{% end if %}

	{% if .LoggedIn %}
		Yep!
	{% else %}
		Negative: No one is logged in.
	{% end if %}

Statement - With

With takes the specified selector and roots a sub-context at that position in
the context.

	{% with .Selector %}...{% end with %}

	{% with .LoggedInUser %}
		Hello {% .FristName %},
		How are you Ms. {% .LastName %}
	{% end with %}

	{% with /. %}
		Now we're rooted back at the top level no matter what!
	{% end with %}

	{% with .LoggedInUser %}
		{% with .FirstName %}
			Hello {% . %},
			How are you Ms. {% $.LastName %}
		{% end with %}
	{% end with %}

Modes

Tmpl has two modes, Production and Development, which can be changed at any time
with the CompileMode function. In Development mode, every block and template is 
loaded from disk and compiled in Execute, so that the latest results are always
used. In Production mode, files are only compiled the first time they are needed
and the results are cached for subsequent access.

Full Implementation Example

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
