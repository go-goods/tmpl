/*
Package tmpl provides a lazy compilation block based templating system.

Tmpl is a simple block based templating system. What does block based mean? When
doing web design typically one thinks of "blocks" of content, for example in a
navigation bar, the main content section of the page, or the header section of
the page. For example, we could have the template defined in "base.tmpl"

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

we could load an execute this template by

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

Insert description of selectors
Insert description of context paths
Insert description of templating language constructs
*/
package tmpl
