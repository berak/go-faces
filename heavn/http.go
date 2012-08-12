package main

import (
	"fmt"
	"math"
	"net/http"
    "sort"
    //~ "strconv"
	
    "appengine"
    "appengine/datastore"
	"appengine/memcache"
	
	"io"
	"bytes"
	"encoding/binary"
	//~ "compress/zlib"

	"image"
	"image/jpeg"
	_ "image/color"
	_ "image/gif"
	_ "image/png"
)


type Thumbnail struct {
	Name string
	Pic []byte
}




func handle(m image.Image, r image.Rectangle) (mat *Matrix) {
	i := 0
	mat = createMatrix(r.Max.X-r.Min.X,r.Max.Y-r.Max.Y)
    for y := r.Min.Y; y < r.Max.Y; y++ {
        for x := r.Min.X; x < r.Max.X; x++ {
			r8,g8,b8,_ := m.At(x,y).RGBA()
			mat.e[i] = uint8((r8>>2) + (g8>>1) + (b8>>2))
			i += 1
		}
	}
	return
}

func handleRGBA(m *image.RGBA) (mat *Matrix) {
	i := 0
    r := m.Bounds()
	mat = createMatrix(r.Max.X,r.Max.Y)
    for y := r.Min.Y; y < r.Max.Y; y++ {
        pix := m.Pix[y*m.Stride:]
        for x := r.Min.X; x < r.Max.X; x++ {
            r8 := uint8(pix[x*4])
            g8 := uint8(pix[x*4+1])
            b8 := uint8(pix[x*4+2])
			mat.e[i] = uint8((r8>>2) + (g8>>1) + (b8>>2))
			i += 1
		}
	}
	return
}
func handleYCbCr(m *image.YCbCr)  (mat *Matrix) {
	i := 0
    r := m.Bounds()
	mat = createMatrix(r.Max.X,r.Max.Y)
    for y := r.Min.Y; y < r.Max.Y; y++ {
        for x := r.Min.X; x < r.Max.X; x++ {
			r8,g8,b8,_ := m.At(x,y).RGBA()
			mat.e[i] = uint8((r8>>2) + (g8>>1) + (b8>>2))
			i += 1
		}
	}
	return;
}



func fillDict(c appengine.Context) {
    var pr []*PicRec
    q := datastore.NewQuery("PicRec")
    _,err := q.GetAll(c,&pr)
    if err != nil {
        c.Warningf("%v",err)
        return
    }
	persons = make(map[string]*Histogram)
    for _, p := range(pr) {
		hist := make(Histogram,len(p.Hist)/2)
		buf  := bytes.NewBuffer(p.Hist)
		binary.Read(buf, binary.LittleEndian, hist)
		persons[p.Name] = &hist
    }
	c.Infof("loaded %v persons",len(persons) )
}

type KSarr []string 
func (s KSarr) Len() int 		   { return len(s) }
func (s KSarr) Less(i, j int) bool { return s[i] < s[j] }
func (s KSarr) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }

func listDict() string {
	var s string = "<table>"
	i := 0
	lp := len(persons)
	arr := make(KSarr,lp)	
    for a, _ := range(persons) {
		arr[i]=a;
		if i+=1; i>=1000 { break }
	}
	sort.Sort(arr)
    for i=0; i<len(arr); i++ {
		if i % 12 == 0 {
			s += "</tr>\n<tr>"
		}
		s += fmt.Sprintf("<td><a href=/?match=%v>%v</a>&nbsp;&nbsp; </td>",arr[i],arr[i])
    }
	return s + "\n</tr></table>\n"
}



var style string = "<style>.z{font-family: Monospace}\n a{text-decoration:none; color:#aaa; font-size:11;}\n a:hover{color:#ddd}\n body,table,pre,input{font-family:Arial,'MS Trebuchet',sans-serif; font-size:12; background-color:#292929; color:#aaa}\n input,file,button{border-color:#777; border-style:solid; border-width:1}\n body{margin: 15 15 15 15}</style>\n"
var persons map[string]*Histogram


func init() {
	initIndices()
	
	http.HandleFunc("/", handler)
	http.HandleFunc("/thumb", handler_thumb)
	http.HandleFunc("/reset", handler_reset)
}


func handler_thumb(w http.ResponseWriter, r *http.Request) {
    r.ParseForm()
    c := appengine.NewContext(r)
	name := r.FormValue("id")

	var thumb Thumbnail
	if item, err := memcache.Get(c, name); err == nil {
		thumb = Thumbnail{Name:item.Key, Pic:item.Value}
	} else {
		q := datastore.NewQuery("Thumbnail").Filter("Name =", name )
		for t := q.Run(c); ; {
			_, err := t.Next(&thumb)
			if err != nil {
				fmt.Fprint(w, "<html><head> ", style, "<title>;(</title></head><body>")
				fmt.Fprint(w, "<h4>can't find thumb for ", name)
				return
			}

			// Add the item to the memcache, if the key does not already exist
			item := &memcache.Item{
				Key:   name,
				Value: thumb.Pic,
			}
			if err := memcache.Add(c, item); err == memcache.ErrNotStored {
				c.Warningf("item with key %q already exists", item.Key)
			} else if err != nil {
				c.Warningf("error adding item: %v", err)
			}
			break
		}
	}

	if thumb.Pic == nil {
		fmt.Fprint(w, "<html><head> ", style, "<title>;(</title></head><body>")
		fmt.Fprint(w, "<h4>can't find thumb for ", name)
		return
	}

    w.Header().Set("Content-Type", "image/jpeg")
    w.Header().Set("Content-Length", fmt.Sprintf("%v",len(thumb.Pic)) )

	io.Copy( w, bytes.NewBuffer(thumb.Pic) )
}


func handler_reset(w http.ResponseWriter, r *http.Request) {
	persons = nil
	http.Redirect(w, r, "/", http.StatusFound)
}


// sortable tmp array of recognition results
type KV struct {
	p string
    d float64
}
type KVarr []KV 
func (s KVarr) Len() int 		   { return len(s) }
func (s KVarr) Less(i, j int) bool { return s[i].d < s[j].d }
func (s KVarr) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }

func compare(match string, hist *Histogram) (compres string) {
	if hist==nil { return }
	if len(persons)<2 { return }
	compres = fmt.Sprintf("&nbsp;&nbsp;<img src='/thumb?id=%v'> &nbsp;&nbsp; matches for : <i>%v</i> <br><br>\n",match,match)
	kv := make(KVarr,len(persons)-1)
	i := 0
	mm := 0.0
	for n,h := range persons {
		if match == n { continue }
		dist := norml2(h,hist)
		kv[i] = KV{p:n,d:dist}
		mm = math.Max(mm,dist)
		i += 1
	}
	sort.Sort(kv)
	for i=0; i<int(math.Min(10.0,float64(len(kv)))); i++ {
		confidence := 1.0-(kv[i].d/mm)
		zz:=""; 
		if confidence > 0.59 { zz = "style='border-style:solid;'"}
		compres += fmt.Sprintf("&nbsp;&nbsp;<a href='/?match=%v' title='%v : %3.3f'><img src='/thumb?id=%v' %v></a>\n", kv[i].p, kv[i].p, confidence, kv[i].p, zz) 
	}
	return 
}

func handler(w http.ResponseWriter, r *http.Request) {
    c := appengine.NewContext(r)
    r.ParseMultipartForm(1000000)
	if persons == nil { fillDict(c) }
	hr := ""
	var m *Matrix = nil
	var hist *Histogram = nil
	
	name := r.FormValue("n");
	if _,e := persons[name]; e {
		hist = persons[name]
	} else {
		f, fhead, err := r.FormFile("f")
		if err == nil {
			defer f.Close()
			if name=="" { name = fmt.Sprintf("%v",fhead)[2:12] }
			// Grab the image data
			var buf bytes.Buffer
			io.Copy(&buf, f)
			im, imtype, err := image.Decode(&buf)
			if err == nil {
				b := im.Bounds()
				//~ if v:=r.FormValue("x"); v!="" { if x,e := strconv.Atoi(v); e==nil { b.Min.X=x } }
				//~ if v:=r.FormValue("y"); v!="" { if x,e := strconv.Atoi(v); e==nil { b.Min.Y=x } }
				//~ if v:=r.FormValue("w"); v!="" { if x,e := strconv.Atoi(v); e==nil { b.Max.X=b.Min.X + x } }
				//~ if v:=r.FormValue("h"); v!="" { if x,e := strconv.Atoi(v); e==nil { b.Max.Y=b.Min.Y + x } }
				hr = fmt.Sprintf("%v %v %v", name,imtype, b)
				c.Infof(hr)
				im = Resize(im, b, 90, 90)
				switch it := im.(type) {
				case *image.RGBA:
					m = handleRGBA(im.(*image.RGBA))
					break
				case *image.YCbCr:
					m = handleYCbCr(im.(*image.YCbCr))
					break
				default:
					c.Warningf("can't resolve image type %v %v", it,imtype)
				}
				if m != nil {
					hist = m.histogram(circle, 1)
					persons[name] = hist
				}
				// it's a new one, so save a thumbnail
				buf.Reset()
				jpeg.Encode(&buf, Resize(im, im.Bounds(), 50, 50), nil)
				thumb := Thumbnail{ Name:name, Pic:buf.Bytes()}
				datastore.Put(c, datastore.NewIncompleteKey(c, "Thumbnail", nil), &thumb)  						
			} else {
				c.Warningf("%v",err)
			}
		} else {
			c.Warningf("%v",err)
		}
	}


	compres := ""
	if hist != nil {
		compres = compare(name, hist)
		
		buf := new(bytes.Buffer)
		rr  := binary.Write(buf, binary.LittleEndian, hist)
		if rr != nil {
			c.Warningf("%v",rr)
		}
		nb := buf.Bytes()

			//~ var b bytes.Buffer
			//~ w := zlib.NewWriter(&b)
			//~ w.Write(buf.Bytes())
			//~ w.Close()
			//~ nb = b.Bytes()
			
		p := PicRec { Name: name, Hist: nb }
		_, err := datastore.Put(c, datastore.NewIncompleteKey(c, "PicRec", nil), &p)    
		if err != nil {
			c.Warningf("%v",err)
		}
		
	} else if match := r.FormValue("match"); match!="" { 
		compres = compare(match,persons[match])
	}
	
	
    w.Header().Add("Content-Type", "text/html")
    fmt.Fprint(w, "<html><head> ", style, "<title></title></head><body>\n")
    //~ fmt.Fprint(w, "<form method=POST enctype='multipart/form-data' onSubmit='postCanvasToURL()'>\n")
    //~ fmt.Fprint(w, "<form action='#' onSubmit='postCanvasToURL()'>\n")
    fmt.Fprint(w, "<table width='80%'><tr><td><table><tr><td><input size=40 id='n' name='n' title='your name here'></td>\n")
    //~ fmt.Fprint(w, "<input size=90 name=i title='your image url here'><p>\n")
    fmt.Fprint(w, "<td><input type=button value='Upload and match this image' id='submit' style='visibility:hidden;' onClick='postCanvasToURL()'></td></tr></table><p>\n")
    //~ fmt.Fprint(w, "<td><input type=button value='Upload and match this image' id='submit' style='visibility:hidden;' onClick='upload()'></td></tr></table><p>\n")
    fmt.Fprint(w, "<input size=60 name=f id='f' title='upload from disk' type=file onChange='loadim()'><p></td>\n")
    fmt.Fprint(w, "<td><div id=compout><br></div></td><td><canvas id='can' width=90 height=90></td></tr></table>\n")
    //~ fmt.Fprint(w, "<input type=hidden name='x'><p>\n")
    //~ fmt.Fprint(w, "<input type=hidden name='y'><p>\n")
    //~ fmt.Fprint(w, "<input type=hidden name='w'><p>\n")
    //~ fmt.Fprint(w, "<input type=hidden name='h'><p>\n")
    fmt.Fprint(w, "<p><hr NOSHADE>\n")
    fmt.Fprint(w, compres, "<p><hr NOSHADE>", listDict(),"<p><hr NOSHADE>" )
	fmt.Fprint(w, "<script type='text/javascript' src='js/ccv.js'></script>\n")
	fmt.Fprint(w, "<script type='text/javascript' src='js/face.js'></script>\n")
	fmt.Fprint(w, "<script type='text/javascript' src='js/faceup.js'></script>")
    fmt.Fprintf(w, "%v\n", hr)
    fmt.Fprint(w, "</body></html>")
}