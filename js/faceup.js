canvas = document.getElementById("can");
ctx = canvas.getContext("2d");
ctx.fillRect(0,0,90,90)


if (XMLHttpRequest.prototype.sendAsBinary === undefined) {
  XMLHttpRequest.prototype.sendAsBinary = function(string) {
    var bytes = Array.prototype.map.call(string, function(c) {
      return c.charCodeAt(0) & 0xff;
    });
    this.send(new Uint8Array(bytes).buffer);
  };
}


function postCanvasToURL() {
	document.getElementById("submit").style.visibility = "hidden"
	var type = "image/jpeg"
	var data = canvas.toDataURL(type);
	if ( ! data ) return
	data = data.replace('data:' + type + ';base64,', '');

	var xhr = new XMLHttpRequest();
	xhr.open('POST', "/", true);
	xhr.onreadystatechange = function()	{
		if ( xhr.readyState < 4 )
			document.getElementById("compout").innerHTML += xhr.readyState + " " + xhr.status + " " + "posting ..<br>"
		else
			document.innerHTML = xhr.responseText
	}
	var fn = document.getElementById("f").value
	var boundary = 'ohaiimaboundary';
	xhr.setRequestHeader('Content-Type', 'multipart/form-data; boundary=' + boundary);
	xhr.sendAsBinary([
		'--' + boundary,
		'Content-Disposition: form-data; name="f"; filename="' + fn + '"',
		'Content-Type: ' + type,
		'',
		atob(data),
		'--' + boundary,
		''
	].join('\r\n'));
}


function loadim() {
	document.getElementById("submit").style.visibility = "hidden"
	var matchim = document.getElementById("f")
	var co = document.getElementById("compout")
	var image = new Image();
	image.src = matchim.files[0].getAsDataURL()
	image.onload = function () {
		ctx.fillRect(0,0,90,90)

		var comp = ccv.detect_objects({ "canvas" : ccv.grayscale(ccv.pre(image)),
										"cascade" : cascade,
										"interval" : 5,
										"min_neighbors" : 1 });
										
		co.innerHTML = comp.length  + " faces found.<br>"
		if( comp.length  > 0 ) {
			var x = Math.ceil(comp[0].x)
			var y = Math.ceil(comp[0].y)
			var w = Math.ceil(comp[0].width)
			var h = Math.ceil(comp[0].height)
			if ( h + h/4 < 90 ) {
				h += h/4
			}
			co.innerHTML += "[" + x + "," + y + "," + w + "," + h + "]<br>\n"
			ctx.drawImage(image, x,y,w,h, 0,0,90,90);
			document.getElementById("submit").style.visibility = "visible"
		}
	}	
	image.onerror = function (e,url,line) {
		alert("eerrr " + e.message + " : " + line + " : " + url)
	}
	return true;
}
