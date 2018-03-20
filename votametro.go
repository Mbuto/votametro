package votametro

import (
    "fmt"
    "net/http"
    "appengine"
    "appengine/datastore"
    "appengine/user"
//    "strconv"
    "time"
   "strings"
    "regexp"
    "math/rand"
    "encoding/xml"
)

const version = "1.3"

const maxvoci = 10

type Area struct {
Usr string
Area_id string
Nome string
Descriz string
Stato string
}

type Opzione struct {
Usr string	`xml:"-"`
Area_id string	`xml:"-"`
Stringa string	`xml:"id,attr"`
Voti int
}

type Antispam struct {
Usr string
Area_id string
Datetime string
}

func init() {
    http.HandleFunc("/", root)
    http.HandleFunc("/miovota", miovota)
    http.HandleFunc("/newarea", newarea)
    http.HandleFunc("/newdo", newdo)
    http.HandleFunc("/visarea", visarea)
    http.HandleFunc("/cancarea", cancarea)
    http.HandleFunc("/votazione", votazione)
    http.HandleFunc("/dovoto", dovoto)
    http.HandleFunc("/attiva", attiva)
    http.HandleFunc("/disattiva", disattiva)
    http.HandleFunc("/lista", lista)
    http.HandleFunc("/visualizza", visualizza)
    http.HandleFunc("/doxml.xml", doxml)
    http.HandleFunc("/help", help)
    http.HandleFunc("/robots.txt", robots)
}

func root(w http.ResponseWriter, r *http.Request) {

c := appengine.NewContext(r)
//if user.Current(c) == nil {
// http.Error(w,"Invalid User",500)
// return
//}

    fmt.Fprintf(w, mioForm1,user.Current(c))
    fmt.Fprintf(w, botForm , version)

}

func miovota(w http.ResponseWriter, r *http.Request) {
const maxarea = 5
var area []Area

c := appengine.NewContext(r)
if user.Current(c) == nil {
 http.Error(w,"Invalid User",500)
 return
}

  fmt.Fprintf(w, mioForm2,user.Current(c))

area = nil
qq := datastore.NewQuery("Area").Filter("Usr =", user.Current(c).String()).Order("Nome")
_, err6 := qq.GetAll(c, &area)
        if err6 != nil {
                fmt.Fprintf(w, "err6=%v\n",err6)
                return
        }

 for k:=0; k<len(area); k++ {
  fmt.Fprintf(w, mioForm3,area[k].Area_id,
      area[k].Nome,area[k].Descriz,area[k].Stato,"Dettagli",area[k].Area_id)
 }

if len(area) < maxarea {
 fmt.Fprintf(w, mioForm4)
} else {
 fmt.Fprintf(w, mioForm5,maxarea)
}

}

func newarea(w http.ResponseWriter, r *http.Request) {

c := appengine.NewContext(r)
if user.Current(c) == nil {
 http.Error(w,"Invalid User",500)
 return
}

  fmt.Fprintf(w, mioArea1,user.Current(c))

 for k:=0; k<maxvoci; k++ {
  fmt.Fprintf(w, mioArea2,k,k)
 }

  fmt.Fprintf(w, mioArea3, maxvoci)
}

func newdo(w http.ResponseWriter, r *http.Request) {
var nome_area string
var descriz string
var voce [maxvoci]string
var p string
var nvoci int

c := appengine.NewContext(r)
if user.Current(c) == nil {
 http.Error(w,"Invalid User",500)
 return
}

    nome_area=r.FormValue("nm")
    if nome_area == "" || check_irreg(w,nome_area,0) {
	visualizza_errore(w,"nome_area",
	"Il nome dell'Area deve essere composto da caratteri alfabetici o numerici, e non pu&ograve; contenere spazi n&eacute; altri caratteri speciali.")
	return
    }

    descriz=r.FormValue("de")
    if descriz == "" || check_irreg(w,descriz,1) {
	visualizza_errore(w,"descrizione",
	"La descrizione deve essere composta da caratteri alfabetici o numerici, e non pu&ograve; contenere caratteri speciali. Pu&ograve; contenere spazi, non all'inizio.")
	return
    }

nvoci = 0
for k:=0; k<len(voce); k++ {
    p = fmt.Sprintf("v%d",k)
    voce[k]=strings.TrimRight(r.FormValue(p)," ")
    if check_irreg(w,voce[k],1) {
	visualizza_errore(w,p,
	"Ogni voce deve essere composta da caratteri alfabetici o numerici, e non pu&ograve; contenere caratteri speciali. Pu&ograve; contenere spazi.")
	return
     }
 for j:=0; j<k; j++ {
	if voce[j] != "" && voce[j] == voce[k] {
//	fmt.Fprintf(w,"voce duplicata %d %d",k,j)
	q := fmt.Sprintf("voce duplicata: Voce %d == Voce %d",k,j)
	visualizza_errore(w,q,
	"Ogni voce deve essere differente da tutte le altre voci.")
	return
	}
 }
 if voce[k] != "" {
   nvoci++
 }
}

if nvoci < 2 {
	visualizza_errore(w,"numero voci valide < 2",
	"Devono essere definite almeno 2 voci.")
	return
}

 fmt.Fprintf(w,"<html><head><link rel=icon href=/favicon.ico /><title>Votametro&trade;</title><style>.grn {background-color: lime}</style></head><body><h1>Votametro!</h1><h3>Risultati dell'inserimento</h3><ul><li>Area: %s %s<ul>\n",nome_area,descriz)
 for k:=0; k<len(voce); k++ {
  fmt.Fprintf(w,"<li>Voce %d: %s\n",k,voce[k])
}
rnd := rand.New(rand.NewSource(time.Now().Unix()))
rndint := rnd.Int63()
p = fmt.Sprintf("%d",rndint)
fmt.Fprintf(w, "</ul><li>area_id=%s\n",p)
fmt.Fprintf(w, "<li>owner=%s\n",user.Current(c).String())

uu := Area{user.Current(c).String(),p,nome_area,descriz,"C"}
_, err5 := datastore.Put(c, datastore.NewIncompleteKey(c, "Area", nil), &uu)
    if err5 != nil {
        fmt.Fprintf(w, "<p>err5=%v\n",err5)
        return
    }
fmt.Fprintf(w, "<li class=grn>Area inserita: %s %s\n<ol>",uu.Nome,uu.Descriz)
 for k:=0; k<len(voce); k++ {
   if (voce[k] != "") {
  vv := Opzione{user.Current(c).String(),p,voce[k],0}
_, err6 := datastore.Put(c, datastore.NewIncompleteKey(c, "Opzione", nil), &vv)
    if err6 != nil {
        fmt.Fprintf(w, "<p>err6=%v\n",err6)
        return
    }
   fmt.Fprintf(w,"<li class=grn>Voce %d inserita: %s\n",k,vv.Stringa)
   }
  }
//    http.Redirect(w,r,"/miovota",302)
fmt.Fprintf(w,"</ol></ul><p><a href=/miovota>Ok</a></body></html>")
}

func visualizza_errore(w http.ResponseWriter, err string, expl string) {
fmt.Fprintf(w,errForm,err,expl)
}

func visarea(w http.ResponseWriter, r *http.Request) {
var area_id string
var opzione []Opzione
var area []Area

c := appengine.NewContext(r)
if user.Current(c) == nil {
 http.Error(w,"Invalid User",500)
 return
}

    area_id=check_num(w,r.FormValue("aa"))

area = nil
rr := datastore.NewQuery("Area").Filter("Usr =", user.Current(c).String()).Filter("Area_id =",area_id).Order("Nome").Limit(1)
_, err5 := rr.GetAll(c, &area)
        if err5 != nil {
                fmt.Fprintf(w, "err5=%v\n",err5)
                return
        }

if area != nil {
  fmt.Fprintf(w, aaForm2,user.Current(c),area[0].Nome,area[0].Descriz,
	area_id,area[0].Stato)
} else {
                fmt.Fprintf(w, "General Sputtanation Error: %s\n",area_id)
                return
}

opzione = nil
qq := datastore.NewQuery("Opzione").Filter("Usr =", user.Current(c).String()).Filter("Area_id =",area_id).Order("Stringa")
_, err6 := qq.GetAll(c, &opzione)
        if err6 != nil {
                fmt.Fprintf(w, "err6=%v\n",err6)
                return
        }

 for k:=0; k<len(opzione); k++ {
  fmt.Fprintf(w, aaForm3,opzione[k].Stringa,opzione[k].Voti)
 }

fmt.Fprintf(w, aaForm4,area_id)

switch area[0].Stato {
case "C":
	fmt.Fprintf(w, aaForm4att, area_id)
	fmt.Fprintf(w, aaForm4canc, area_id)
case "A":
	fmt.Fprintf(w, aaForm4dis, area_id)
}
fmt.Fprintf(w, aaForm4prefine,area_id,area_id)
fmt.Fprintf(w, aaForm4fine)

}

func cancarea(w http.ResponseWriter, r *http.Request) {
var area_id string
var bruto string
var area []Area

c := appengine.NewContext(r)
if user.Current(c) == nil {
 http.Error(w,"Invalid User",500)
 return
}

    area_id=check_num(w,r.FormValue("aa"))
    bruto=r.FormValue("bruto")

if bruto == "" {

area = nil
rr := datastore.NewQuery("Area").Filter("Usr =", user.Current(c).String()).Filter("Area_id =",area_id).Order("Nome").Limit(1)
_, err5 := rr.GetAll(c, &area)
        if err5 != nil {
                fmt.Fprintf(w, "err5=%v\n",err5)
                return
        }

if area != nil {
  fmt.Fprintf(w, cancForm1,user.Current(c),area[0].Nome,area[0].Descriz,
	area_id,area[0].Stato,area_id)
} else {
                fmt.Fprintf(w, "General Sputtanation Error: %s\n",area_id)
                return
}

} else {
fmt.Fprintf(w, cancForm2,user.Current(c),area_id)

delall(w,r,c,"Area",area_id)
delall(w,r,c,"Opzione",area_id)

fmt.Fprintf(w, cancForm3,area_id)
}

}


func delall(w http.ResponseWriter, r *http.Request, c appengine.Context, tab string, area_id string) {

q := datastore.NewQuery(tab).Filter("Usr =", user.Current(c).String()).Filter("Area_id =",area_id).Limit(100).KeysOnly()

for  {

zzz2, err9 := q.GetAll(c, nil)
if err9 != nil {
	fmt.Fprintf(w, "GetAll err9=%v\n",err9)
return
}

if len(zzz2) == 0 {
	break
}

err8 := datastore.DeleteMulti(c, zzz2)
if err8 != nil {
	fmt.Fprintf(w, "<p>DeleteMulti err8=%v\n",err8)
return
}

}

}


func check_num(w http.ResponseWriter, c string) string {
const goodsetnum = "^([0-9])*$"
var d *regexp.Regexp
var reerr error

d, reerr = regexp.Compile(goodsetnum)

if (reerr != nil) {
fmt.Fprintf(w, "Regexp compile error: %x", reerr)
return ""
}
if (d.MatchString(c)) { return c }
return ""
}

func check_irreg(w http.ResponseWriter, c string, f int) bool {
const goodset1 = "^([0-9A-Za-z_-àèéìòù])*$"
const goodset2 = "^([ 0-9A-Za-z_-àèéìòù])*$"
var d *regexp.Regexp
var reerr error

if (f == 0) {
d, reerr = regexp.Compile(goodset1)
} else {
if strings.HasPrefix(c," ") { return true }
d, reerr = regexp.Compile(goodset2)
}
if (reerr != nil) {
fmt.Fprintf(w, "Regexp compile error: %x", reerr)
return false
}
if (d.MatchString(c)) { return false }
return true
}

func votazione(w http.ResponseWriter, r *http.Request) {
var area_id string
var fake string
var action string
var comment string
var opzione []Opzione
var area []Area

c := appengine.NewContext(r)
if user.Current(c) == nil {
 http.Error(w,"Invalid User",500)
 return
}
//        if r.Method != "POST" {
//                http.NotFound(w, r)
//                return
//        }


    area_id=check_num(w,r.FormValue("aa"))
    fake=r.FormValue("fake")

if fake == "" {
 if check_attiva(c, w, area_id) > 0 {
  fmt.Fprintf(w, areaInattiva, user.Current(c).String(), area_id)
  return
 }

 if check_voto(c, w, user.Current(c).String(), area_id) > 0 {
  fmt.Fprintf(w, haiGiaVotato, user.Current(c).String(), area_id)
  return
 }
}

area = nil
rr := datastore.NewQuery("Area").Filter("Area_id =",area_id).Order("Nome").Limit(1)
_, err5 := rr.GetAll(c, &area)
        if err5 != nil {
                fmt.Fprintf(w, "err5=%v\n",err5)
                return
        }

if area != nil {
  if fake == "" {
    action = "/dovoto"
    comment = ""
  } else {
    action = "/"
    comment = "(voto di prova: questo voto non verr&agrave; registrato)"
  }
  fmt.Fprintf(w, votoForm2, user.Current(c), area[0].Nome, area_id,
	area[0].Stato, area[0].Descriz, action)
} else {
                fmt.Fprintf(w, "General Sputtanation Error: %s\n",area_id)
                return
}

opzione = nil
qq := datastore.NewQuery("Opzione").Filter("Area_id =",area_id).Order("Stringa")
_, err6 := qq.GetAll(c, &opzione)
        if err6 != nil {
                fmt.Fprintf(w, "err6=%v\n",err6)
                return
        }

 for k:=0; k<len(opzione); k++ {
  fmt.Fprintf(w, votoForm3,opzione[k].Stringa,opzione[k].Stringa)
 }

if check_attiva(c, w, area_id) == 0 {
	fmt.Fprintf(w, votoForm4, area_id, comment)
} else
{
	fmt.Fprintf(w, votoForm5)
}
}

func dovoto(w http.ResponseWriter, r *http.Request) {
var area_id string
var voto string
var opzione []Opzione

c := appengine.NewContext(r)
if user.Current(c) == nil {
 http.Error(w,"Invalid User",500)
 return
}

        if r.Method != "POST" {
                http.NotFound(w, r)
                return
        }

    area_id=check_num(w,r.FormValue("aa"))
    voto=r.FormValue("voto")

if check_attiva(c, w, area_id) > 0 {
 fmt.Fprintf(w, areaInattiva, user.Current(c).String(), area_id)
 return
}

if check_voto(c, w, user.Current(c).String(), area_id) > 0 {
 fmt.Fprintf(w, haiGiaVotato, user.Current(c).String(), area_id)
 return
}

opzione = nil
qq := datastore.NewQuery("Opzione").Filter("Area_id =",area_id).Filter("Stringa =", voto).Limit(1)
key, err6 := qq.GetAll(c, &opzione)
        if err6 != nil {
                fmt.Fprintf(w, "err6=%v\n",err6)
                return
        }

if len(opzione) == 0 {
                fmt.Fprintf(w, "LEN0 err6=%v\n",err6)
                return
}

opzione[0].Voti++

err8 := datastore.DeleteMulti(c, key)
if err8 != nil {
	fmt.Fprintf(w, "DeleteMulti err8=%v\n",err8)
return
}

_, err7 := datastore.Put(c, datastore.NewIncompleteKey(c, "Opzione", nil), &(opzione[0]))
    if err7 != nil {
        fmt.Fprintf(w, "err7=%v\n",err7)
        return
    }

fmt.Fprintf(w, votoInserito, user.Current(c).String(), opzione[0].Stringa,area_id,area_id)

log_voto(c, w, user.Current(c).String(), area_id)

}

func check_voto(c appengine.Context, w http.ResponseWriter, usr string, aid string) int {
var as []Antispam

as = nil
qq := datastore.NewQuery("Antispam").Filter("Usr =",usr).Filter("Area_id =",aid)
_, err6 := qq.GetAll(c, &as)
        if err6 != nil {
                fmt.Fprintf(w, "err6=%v\n",err6)
                return 2
        }
if len(as) > 0 {
  return 1
}
return 0
}

func log_voto(c appengine.Context, w http.ResponseWriter, usr string, aid string) {
uu := Antispam{user.Current(c).String(),aid,time.Now().String()}
_, err5 := datastore.Put(c, datastore.NewIncompleteKey(c, "Antispam", nil), &uu)
    if err5 != nil {
        fmt.Fprintf(w, "err5=%v\n",err5)
        return
    }
}

func check_attiva(c appengine.Context, w http.ResponseWriter, aid string) int {
var ar []Area

ar = nil
qq := datastore.NewQuery("Area").Filter("Area_id =",aid)
_, err6 := qq.GetAll(c, &ar)
        if err6 != nil {
                fmt.Fprintf(w, "err6=%v\n",err6)
                return 2
        }
if len(ar) == 0 {
  return 1
}
if ar[0].Stato != "A" {
return 1
}
return 0
}

func attiva(w http.ResponseWriter, r *http.Request) {
var area_id string

c := appengine.NewContext(r)
if user.Current(c) == nil {
 http.Error(w,"Invalid User",500)
 return
}
    area_id=check_num(w,r.FormValue("aa"))
    cambia_stato(c, w, area_id, "A")
    http.Redirect(w,r,"/",302)
}

func disattiva(w http.ResponseWriter, r *http.Request) {
var area_id string

c := appengine.NewContext(r)
if user.Current(c) == nil {
 http.Error(w,"Invalid User",500)
 return
}
    area_id=check_num(w,r.FormValue("aa"))
    cambia_stato(c, w, area_id, "C")
    http.Redirect(w,r,"/",302)
}

func cambia_stato(c appengine.Context, w http.ResponseWriter, aid string, sts string) {
var area []Area

area = nil
qq := datastore.NewQuery("Area").Filter("Area_id =",aid)
key, err6 := qq.GetAll(c, &area)
        if err6 != nil {
                fmt.Fprintf(w, "err6=%v\n",err6)
                return
        }

area[0].Stato = sts

err8 := datastore.DeleteMulti(c, key)
if err8 != nil {
	fmt.Fprintf(w, "DeleteMulti err8=%v\n",err8)
return
}

_, err7 := datastore.Put(c, datastore.NewIncompleteKey(c, "Area", nil), &(area[0]))
    if err7 != nil {
        fmt.Fprintf(w, "err7=%v\n",err7)
        return
    }
}

func lista(w http.ResponseWriter, r *http.Request) {

var votabile string
var link string
var class string
var area []Area

c := appengine.NewContext(r)
if user.Current(c) == nil {
 http.Error(w,"Invalid User",500)
 return
}

area = nil
qq := datastore.NewQuery("Area").Filter("Stato =","A")
_, err6 := qq.GetAll(c, &area)
        if err6 != nil {
                fmt.Fprintf(w, "err6=%v\n",err6)
                return
        }
fmt.Fprintf(w, fmLista1, user.Current(c).String())
if len(area) == 0 {
 fmt.Fprintf(w, "<tr><th colspan=3 class=alrt>Non ci sono Aree disponibili per la votazione.</th></tr>")
} else {
for k:=0; k<len(area); k++ {
	if check_voto(c, w, user.Current(c).String(), area[k].Area_id) == 0 {
		votabile = "PUOI VOTARE"
		link = fmt.Sprintf("<a href=/votazione?aa=%s>", area[k].Area_id)
		class = "grn"
	} else {
		votabile = "HAI GI&Agrave; VOTATO"
		link = ""
		class = "alrt"
	}
	link2 := fmt.Sprintf("<a href=/visualizza?aa=%s>", area[k].Area_id)
	fmt.Fprintf(w, fmLista2, class, link, votabile, area[k].Descriz, link2)
}
}
fmt.Fprintf(w, fmLista3)
}

func visualizza(w http.ResponseWriter, r *http.Request) {

var totvoti int
var area_id string
var opzione []Opzione
var area []Area

c := appengine.NewContext(r)
if user.Current(c) == nil {
 http.Error(w,"Invalid User",500)
 return
}

    area_id=check_num(w,r.FormValue("aa"))

area = nil
rr := datastore.NewQuery("Area").Filter("Area_id =",area_id).Order("Nome").Limit(1)
_, err5 := rr.GetAll(c, &area)
        if err5 != nil {
                fmt.Fprintf(w, "err5=%v\n",err5)
                return
        }

if area != nil {
  fmt.Fprintf(w, aaForm2,user.Current(c),area[0].Nome,area[0].Descriz,
	area_id,area[0].Stato)
} else {
                fmt.Fprintf(w, "General Sputtanation Error: %s\n",area_id)
                return
}

totvoti = 0
opzione = nil
qq := datastore.NewQuery("Opzione").Filter("Area_id =",area_id).Order("-Voti")
_, err6 := qq.GetAll(c, &opzione)
        if err6 != nil {
                fmt.Fprintf(w, "err6=%v\n",err6)
                return
        }

grLab := ""
grVal := ""
 for k:=0; k<len(opzione); k++ {
  totvoti += opzione[k].Voti
  fmt.Fprintf(w, aaForm3,opzione[k].Stringa,opzione[k].Voti)
  if grLab == "" {
    grLab = opzione[k].Stringa
  } else {
    grLab = grLab + "|" + opzione[k].Stringa
  }
  if grVal == "" {
    grVal = fmt.Sprintf("%d",opzione[k].Voti)
  } else {
    grVal = grVal + fmt.Sprintf(",%d",opzione[k].Voti)
  }
 }

 lnk := ""
 if check_voto(c, w, user.Current(c).String(), area_id) > 0 {
  lnk = ""
 } else {
  lnk = fmt.Sprintf("<a href=/votazione?aa=%s>Vota</a>",area_id)
 }
fmt.Fprintf(w, vvForm4,totvoti,area_id,area_id,lnk,grLab,grVal)
fmt.Fprintf(w, aaForm4fine)
}

func doxml(w http.ResponseWriter, r *http.Request) {

var area_id string
var opzione []Opzione
var area []Area

c := appengine.NewContext(r)
//if user.Current(c) == nil {
// http.Error(w,"Invalid User",500)
// return
//}

    area_id=check_num(w,r.FormValue("aa"))

area = nil
rr := datastore.NewQuery("Area").Filter("Area_id =",area_id).Order("Nome").Limit(1)
_, err5 := rr.GetAll(c, &area)
        if err5 != nil {
                fmt.Fprintf(w, "err5=%v\n",err5)
                return
        }

if area == nil {
                fmt.Fprintf(w, "General Sputtanation Error: %s\n",area_id)
                return
}

opzione = nil
qq := datastore.NewQuery("Opzione").Filter("Area_id =",area_id).Order("-Voti")
_, err6 := qq.GetAll(c, &opzione)
        if err6 != nil {
                fmt.Fprintf(w, "err6=%v\n",err6)
                return
        }

  output2, err2 := xml.MarshalIndent(opzione, "  ", "    ")
  if err2 != nil {
		fmt.Fprintf(w, "error: %v\n", err2)
		return
  }
  s := fmt.Sprintf("attachment; filename=votametro_%s.xml",area[0].Nome)
  w.Header().Set("Content-Type", "text/xml")
  w.Header().Set("Content-Disposition", s)
  fmt.Fprintf(w, "%s<Votametro>\n<Area id=\"%s\">\n%s\n</Area>\n</Votametro>\n",
	xml.Header,area[0].Nome,output2)

}

func help(w http.ResponseWriter, r *http.Request) {
fmt.Fprintf(w, fmHelp)
}

func robots(w http.ResponseWriter, r *http.Request) {
fmt.Fprintf(w, "User-agent: *\nDisallow: /\n")
}

const fmHelp = `
<html>
  <head>
<link rel="icon" href="/favicon.ico" />
<title>Votametro&trade;</title>
   <style type="text/css">
   .alrt { color: red }
   .grn { background-color: lime }
   .ylw { background-color: yellow }
   .gry { background-color: silver }
   table
   {
   border-collapse:collapse;
   }
   table,th, td
   {
   border: 1px solid black;
   }
  </style>
  </head>
  <body>
  <h1>Votametro!</h1>
  <h3>Requisiti</h3>
Per poter accedere al Votametro&trade; &egrave; necessario avere un Account
Google (<a target=_blank href=https://support.google.com/accounts/answer/27441?hl=it>che cos'&egrave; <img valign=middle src=/images/question.png></a>).<br>
Tutti i link di questa pagina richiedono l'autenticazione Google, se gi&agrave; non effettuata, e si apriranno su pagina nuova, per consentire di tenere questo Help a portata di mano.
  <h3>Help e How-to</h3>
<ul>
<li>Per <b>votare</b>
<ol type=a>
<li>Se sei gi&agrave; in possesso dell'indirizzo (URL) di un'Area di voto: vai a quell'indirizzo e poi segui le istruzioni al <b>punto c</b>.
<li>Se non sei in possesso dell'indirizzo (URL) di un'Area di voto o se vuoi esplorare le diverse Aree disponibili, <a target=_blank href=/lista class=ylw>CLICCA QUI</a> e scegli uno dei link verdi "<span class=grn>PUOI VOTARE</span>", a seconda dell'Area in cui ti interessa votare.<br>
La dicitura "<span class=alrt>HAI GI&Agrave; VOTATO</span>" significa che hai gi&agrave; votato in quell'Area.
<b>Non &egrave; possibile votare pi&ugrave; volte nella stessa Area!</b>
<li>Una volta raggiunta un'Area di voto:
<ol>
<li>Seleziona l'opzione che preferisci fra quelle proposte
<li>Premi il tasto "<code class=gry>Vota!</code>" che troverai sulla pagina.
</ol>
</ol>
<br>
<li>Per vedere i <b>risultati</b> parziali di un'Area di voto:
<ul>
<li>Se sei gi&agrave; in possesso dell'indirizzo (URL) di visualizzazione risultati dell'Area di voto: vai a quell'indirizzo
<li>Se non sei in possesso dell'indirizzo (URL) di visualizzazione risultati dell'Area di voto o se vuoi esplorare le diverse Aree disponibili, <a target=_blank href=/lista class=ylw>CLICCA QUI</a> e scegli uno dei link "<u>VEDI VOTAZIONE</u>"
<li>Una volta raggiunta la pagina di visualizzazione dell'Area di voto:
<br>
<ol>
<li>Se vuoi aggiornare i risultati in tempo reale (voti pervenuti dopo la visualizzazione della pagina), clicca sul link "<code><u>Aggiorna</u></code>" che troverai sulla pagina stessa.
<li>Per uscire, chiudi la pagina o clicca sul link "<code><u>Ok</u></code>" che troverai sulla pagina.
<li>Per votare, clicca sul link "<code><u>Vota</u></code>" che troverai sulla pagina. <b>Questo link comparir&agrave; solo se non hai ancora votato in quell'Area.</b>
</ol>
</ul>
<br>
<li>Se vuoi <b>creare</b> una tua Area di voto personalizzata:
<ul>
<li>Clicca <a target=_blank href=/newarea class=ylw>su questo link</a>
<li>Sulla pagina di creazione di una nuova Area di voto, inserisci:
<ol>
<li>Nome dell'Area: deve contenere solo caratteri alfabetici e numerici, niente spazi
<li>Descrizione dell'Area: deve contenere solo caratteri alfabetici e numerici, pu&ograve; contenere spazi e costituisce la "domanda" a cui la votazione si riferisce
<li>Voce 0 .. Voce 9: da 2 a 10 opzioni di voto (caratteri alfabetici, numerici e spazi), costituiscono le opzioni di voto fra cui i votanti sono chiamati a scegliere
</ol>
<li>Clicca sul tasto <code class=gry>Conferma</code>: saranno eseguiti alcuni controlli sui valori introdotti e, in caso di errori, ne sar&agrave; abilitata la correzione; dopo la correzione, clicca di nuovo sul tasto <code class=gry>Conferma</code>, finch&eacute; non otterrai un risultato corretto, confermato da apposito messaggio. Se invece vuoi abbandonare la creazione dell'Are di voto, clicca sul link "<u>Annulla</u>"
<li>Dopo la creazione, l'Area non &egrave; ancora disponibile al pubblico per la votazione: <b>&egrave; necessario attivarla</b> (oppure la si pu&ograve; cancellare)
<li>&Egrave; possibile creare un massimo di <b>cinque</b> Aree di voto personalizzate. In caso di necessit&agrave;, se ne pu&ograve; cancellare una esistente (dopo averla disattivata).
</ul>
<br>
<li>Se vuoi <b>attivare</b> cio&egrave; rendere disponibile per la votazione un'Area da te creata, fai come segue:
<ul>
<li>Vai a <a target=_blank href=/miovota class=ylw>questa pagina</a>: troverai la tua Area di voto in Stato "<b>C</b>"
<li>Clicca sul bottone <code class=gry>Dettagli</code> relativo all'Area
<li>Preleva il codice HTML per la votazione (puoi salvare il link che compare a fine pagina)
<li>Clicca sul bottone <code class=gry>Attiva questa Area di voto</code>
<li>Tornando a <a target=_blank href=/miovota class=ylw>questa pagina</a>: troverai la tua Area di voto in Stato "<b>A</b>"
<li>Se hai perso l'indirizzo per la votazione, clicca sul bottone <code class=gry>Dettagli</code> relativo all'Area e preleva il codice HTML
</ul>
<br>
<li>Se vuoi <b>disattivare</b> cio&egrave; rendere non disponibile per la votazione un'Area da te creata, fai come segue:
<ul>
<li>Vai a <a target=_blank href=/miovota class=ylw>questa pagina</a>: troverai la tua Area di voto in Stato "<b>A</b>"
<li>Clicca sul bottone <code class=gry>Dettagli</code> relativo all'Area
<li>Clicca sul bottone <code class=gry>Disattiva questa Area di voto</code>
<li>Tornando a <a target=_blank href=/miovota class=ylw>questa pagina</a>: troverai la tua Area di voto in Stato "<b>C</b>"
<li>A questo punto puoi scegliere se <b>cancellarla</b> del tutto usando il bottone <code class=gry>Cancella questa Area di voto</code> e in questo modo perdere tutti i dati di votazione e il grafico, oppure lasciarla inattiva, conservando i dati dalla votazione e il grafico. L'indirizzo per la visualizzazione rimane identico.
</ul>
</ul>
<hr width="50%" align=left>
Altri dubbi o problemi col Votametro&trade;? <img valign=middle src=/favicon.ico width=20 height=20><br>
Contatta l'autore: <i>mbuto11</i> ovviamente su Gmail!
</body></html>
  `

const fmLista1 = `
<html>
  <head>
<link rel="icon" href="/favicon.ico" />
<title>Votametro&trade;</title>
   <style type="text/css">
   .alrt { color: red }
   .grn { background-color: lime }
   table
   {
   border-collapse:collapse;
   }
   table,th, td
   {
   border: 1px solid black;
   }
  </style>
  </head>
  <body>
  <h1>Votametro!</h1>
  <h2>Hello <i>%s</i></h2>
  <h3>Lista delle Aree di voto disponibili:</h3>
  <table><tr>
    <th>Puoi votare?</th>
    <th>Area di voto</th>
    <th>Vai ai risultati<br>parziali della votazione</th>
    </tr>
  `

const fmLista2 = `
<tr><td class=%s>%s%s</a></td><td align=center><b><i>%s</i></b></td><td>%sVEDI VOTAZIONE</a></td></tr>
  `

const fmLista3 = `
</table>
<h3>Vuoi aggiungere una tua Area di voto?</h3>
<a href=/miovota>CLICCA QUI</a><p>ognuno pu&ograve; creare fino a 5 Aree di voto e metterle a disposizione pubblicamente per le votazioni:<br>compariranno  nella lista qui sopra, insieme con le altre Aree gi&agrave; presenti.
<h3>Non vedi un'Area che hai creato?</h3>
<p>Forse devi attivarla.
<a href=/miovota>CLICCA QUI</a>
<h3>Serve aiuto?</h3>
<p>Vuoi maggiori informazioni o indicazioni per fare qualcosa nel Votametro?
<a href=/help>HELP!</a>
<hr width="50%" align=left>
Dubbi o problemi col Votametro&trade;? <img valign=middle src=/favicon.ico width=20 height=20><br>
Contatta l'autore: <i>mbuto11</i> ovviamente su Gmail!
</body></html>
  `

const errForm = `
<html>
  <head>
<link rel="icon" href="/favicon.ico" />
<title>Votametro&trade;</title>
   <style type="text/css">
   .alrt { color: red }
   .wrn { background-color: yellow }
  </style>
  </head>
  <body>
  <h1>Votametro!</h1>
  <h3 class=alrt>Errore %s</h3>
  <p class=wrn>%s
  <p><a href="javascript:history.back(-1)">Correggi</a>
  </body></html>
  `

const areaInattiva = `
<html>
  <head>
<link rel="icon" href="/favicon.ico" />
<title>Votametro&trade;</title>
   <style type="text/css">
   .alrt { color: red }
   .wrn { background-color: yellow }
  </style>
  </head>
  <body>
  <h1>Votametro!</h1>
  <h2>Hello <i>%s</i></h2>
  <h3 class=alrt>L'Area di voto %s &egrave; inattiva</h3>
<a href=/>Ok</a>
  </body></html>
  `

const haiGiaVotato = `
<html>
  <head>
<link rel="icon" href="/favicon.ico" />
<title>Votametro&trade;</title>
   <style type="text/css">
   .alrt { color: red }
   .wrn { background-color: yellow }
  </style>
  </head>
  <body>
  <h1>Votametro!</h1>
  <h2>Hello <i>%s</i></h2>
  <h3 class=alrt>Hai gi&agrave; votato nell'Area: %s</h3>
<a href=/>Ok</a>
  </body></html>
  `

const votoInserito = `
<html>
  <head>
<link rel="icon" href="/favicon.ico" />
<title>Votametro&trade;</title>
   <style type="text/css">
   .alrt { color: red }
   .wrn { background-color: yellow }
  </style>
  </head>
  <body>
  <h1>Votametro!</h1>
  <h2>Hello <i>%s</i></h2>
  <h3 class=wrn>Voto inserito: %s nell'Area: %s</h3>
<a href=/visualizza?aa=%s>Ok</a>
  </body></html>
  `

const mioForm1 = `
<html>
  <head>
<link rel="icon" href="/favicon.ico" />
<title>Votametro&trade;</title>
   <style type="text/css">
   .alrt { background-color: red }
   .wrn { background-color: yellow }
   .gry { background-color: silver }
   table
   {
   border-collapse:collapse;
   }
   table,th, td
   {
   border: 1px solid black;
   }
   </style>
  </head>
  <body>
  <h1>Votametro!</h1>
  <h2>Hello <i>%s</i></h2>
  <h3>Che cos'&egrave; il Votametro&trade;<img valign=middle src=/favicon.ico width=20 height=20></h3>
Il <b>Votametro&trade;</b> permette di votare in una o pi&ugrave;
<i>"aree di voto"</i>
in ciascuna delle quali gli utenti della rete potranno scegliere fra
diverse alternative proposte in voto esclusivo.
  <h3>Come si usa</h3>
<ol type=A>
<li>Per votare
<br>
<ol>
<li>Clicca sul bottone "<code class=gry>Accetto le condizioni e desidero proseguire</code>" in fondo a questa pagina: sar&agrave; visualizzata una lista di Aree di votazione disponibili. Per votare, clicca sul link verde "<code>PUOI VOTARE</code>" dell'Area su cui vuoi votare.
<b>Non &egrave; possibile votare pi&ugrave; volte nella stessa Area!</b>
<li>Seleziona l'opzione che preferisci fra quelle proposte
<li>Premi il tasto "<code class=gry>Vota!</code>" che troverai sulla pagina
<li>Visualizza i risultati in tempo reale: in ogni momento potrai controllare
l'andamento delle votazioni. Il Votametro prevede anche la creazione
di <b>grafici online</b>
</ol>
<br>
<li>Per definire un'area di voto
<br>
<ol>
<li>Definisci la tua <i>area di voto</i>: metti un titolo e una descrizione che spieghi su che cosa si sta votando
<li>Introduci le voci che saranno sottoposte a votazione: introduci le singole voci da votare in alternativa fra loro
<li>Pubblica il codice per dare accesso alla votazione: preleva il codice HTML (che sar&agrave; <b>generato automaticamente</b> dal Votametro) e inseriscilo
nel tuo sito web / blog ecc. per permettere lo svolgimento
pubblico delle votazioni
<li>Visualizza i risultati in tempo reale: in ogni momento potrai controllare
l'andamento delle votazioni. Il Votametro prevede anche la creazione
di <b>grafici online</b>
<li>Chiudi le votazioni: quando ritieni opportuno, puoi consolidare
i dati e impedire che ulteriori voti si aggiungano a quelli gi&agrave;
raccolti. Tutti gli altri strumenti del Votametro (risultati, grafici)
resteranno a tua disposizione finch&eacute; non deciderai di cancellare
l'<i>area di voto</i> con tutti i dati ad essa associati.
</ol>
</ol>
  <h3>Requisiti</h3>
Per poter accedere al Votametro &egrave; necessario avere un Account
Google (<a href=https://support.google.com/accounts/answer/27441?hl=it>che cos'&egrave; <img valign=middle src=/images/question.png></a>)
  <h3>Accesso</h3>
<button onclick="window.location.replace('/lista')">
<b>Accetto le condizioni e desidero proseguire</b></button>
&nbsp;&nbsp;&nbsp;<a href="http://www.google.com/">No, grazie</a>
&nbsp;&nbsp;&nbsp;<a href="/help">Help!</a>
  `

const mioForm2 = `
<html>
  <head>
<link rel="icon" href="/favicon.ico" />
<title>Votametro&trade;</title>
   <style type="text/css">
   .alrt { color: red }
   .wrn { background-color: yellow }
   .cen { margin-left:auto; margin-right:auto; }
   table
   {
   border-collapse:collapse;
   }
   table,th, td
   {
   border: 1px solid black;
   }
   </style>
  </head>
  <body>
  <h1>Votametro!</h1>
  <h2>Hello <i>%s</i></h2>
  <h3>Ecco le tue Aree di voto:</h3>
  <blockquote>
  <table border>
<tr><th>ID</th><th>Nome</th><th>Descrizione</th><th>Stato</th><th>Op.</th></tr>
`

const mioForm3 = `
<tr><td>%s</td><td>%s</td><td>%s</td><th>%s</th>
<td><form action=/visarea method=post><button type=submit>%s</button>
 <input type=hidden name=aa value=%s></form></td>
</tr>
`

const mioForm4 = `
  </table>
  <p><a href=/lista>Visualizza le Aree di voto disponibili</a>
  <p><a href=/newarea>Crea una nuova Area di voto</a>
  </blockquote>
  </body>
</html>
`

const mioForm5 = `
  </table>
  <p class=alrt><b>Hai raggiunto il massimo numero di Aree di voto disponibili (%d)!</b>
  <p>Se vuoi crearne una nuova, prova a cancellarne qualcuna.
  <p><a href=/lista>Visualizza le Aree di voto disponibili</a>
  </blockquote>
  </body>
</html>
`

const aaForm2 = `
<html>
  <head>
<link rel="icon" href="/favicon.ico" />
<title>Votametro&trade;</title>
   <style type="text/css">
   .alrt { background-color: red }
   .wrn { background-color: yellow }
   .cen { margin-left:auto; margin-right:auto; }
   table
   {
   border-collapse:collapse;
   }
   table,th, td
   {
   border: 1px solid black;
   }
   </style>
  </head>
  <body>
  <h1>Votametro!</h1>
  <h2>Hello <i>%s</i></h2>
  <h3>Dettaglio Area di voto: %s - %s (%s %s)</h3>
  <blockquote>
  <table border>
<tr><td>
  <table border>
<tr><th>Opzione</th><th>Voti</th></tr>
`

const votoForm2 = `
<html>
  <head>
<link rel="icon" href="/favicon.ico" />
<title>Votametro&trade;</title>
   <style type="text/css">
   .alrt { color: red }
   .wrn { background-color: yellow }
   .cen { margin-left:auto; margin-right:auto; }
   table
   {
   border-collapse:collapse;
   }
   table,th, td
   {
   border: 1px solid black;
   }
   </style>
  </head>
  <body>
  <h1>Votametro!</h1>
  <h2>Hello <i>%s</i></h2>
  <h3>Area di voto: %s - (%s %s)</h3>
  <h3>%s</h3>
  <blockquote>
  <form action=%s method=post>
  <table border=0>
<tr><th>Opzione</th><th>Voto</th></tr>
`

const cancForm1 = `
<html>
  <head>
<link rel="icon" href="/favicon.ico" />
<title>Votametro&trade;</title>
   <style type="text/css">
   .alrt { color: red }
   .wrn { background-color: yellow }
   .cen { margin-left:auto; margin-right:auto; }
   table
   {
   border-collapse:collapse;
   }
   table,th, td
   {
   border: 1px solid black;
   }
   </style>
  </head>
  <body>
  <h1>Votametro!</h1>
  <h2>Hello <i>%s</i></h2>
  <h3 class=alrt>Vuoi cancellare questa Area di voto: %s - %s (%s %s) ?</h3>
<a href=/miovota>NO Torna indietro</a>
<form action=cancarea method=post>
<p><button type=submit>OK CANCELLA</button>
<input type=hidden name=aa value=%s>
<input type=hidden name=bruto value=yes>
</form>
</body></html>
`


const cancForm2 = `
<html>
  <head>
<link rel="icon" href="/favicon.ico" />
<title>Votametro&trade;</title>
  </head>
  <body>
  <h1>Votametro!</h1>
  <h2>Hello <i>%s</i></h2>
  <h3>Inizio cancellazione Area di voto: %s</h3>
`

const cancForm3 = `
  <h3>Fine cancellazione Area di voto: %s</h3>
<a href=/miovota>Ok</a>
</body></html>
`

const aaForm3 = `
<tr><td>%s</td><td>%d</td>
</tr>
`

const votoForm3 = `
<tr><td>%s</td><td><input type=radio name=voto value="%s" required></td>
</tr>
`

const aaForm4 = `
  </table>
<p><a href=/miovota>Ok</a>&nbsp;&nbsp;
<a href=/votazione?aa=%s&fake=Y>Prova la votazione</a>
`

const vvForm4 = `
  </table>
<p>Totale %d voti
<p><a href=/visualizza?aa=%s>Aggiorna</a>
<p><a href=/lista>Ok</a>
<p><a href=/doxml.xml?aa=%s>XML</a>
<p>%s
</td><td>
<img src="http://chart.apis.google.com/chart?chs=500x250&cht=p&chco=009900|FE6600|990066|0099F0|FE66F0|99F066|F09900|FE0099|6699F0|FE9966&chp=100&chl=%s&chd=t:%s">
</td></tr></table>
`

const aaForm4att = `
<form action=/attiva method=post>
<input type=hidden name=aa value=%s>
<button type=submit>Attiva questa Area di voto</button>
<span onclick="alert('Apri le votazioni')"><img valign=middle src=/images/question.png></span>
</form>
`

const aaForm4dis = `
<form action=/disattiva method=post>
<input type=hidden name=aa value=%s>
<button type=submit>Disattiva questa Area di voto</button>
<span onclick="alert('Chiudi le votazioni')"><img valign=middle src=/images/question.png></span>
</form>
`

const aaForm4canc = `
<form action=/cancarea method=post>
<input type=hidden name=aa value=%s>
<button type=submit>Cancella questa Area di voto</button>
<span onclick="alert('Cancella Area e dati')"><img valign=middle src=/images/question.png></span>
</form>
`

const aaForm4prefine = `
<p><table><tr><th>Codice HTML per questa Area di voto</th></tr>
<tr>
<td>&nbsp;https://votametro.appspot.com/votazione?aa=%s&nbsp;</td>
</tr>
<td>oppure salva <a href=https://votametro.appspot.com/votazione?aa=%s>QUESTO LINK</a></td>
<tr>
</tr></table>
`

const aaForm4fine = `
  </blockquote>
  </body>
</html>
`

const votoForm4 = `
  </table>
<input type=hidden name=aa value=%s>
<p>
<button type=submit>Vota!</button> %s
</form>
<a href=/>Annulla</a>
  </blockquote>
  </body>
</html>
`

const votoForm5 = `
  </table></form>
  <p class=alrt>Questa Area di voto non &egrave; attiva</p>
  </blockquote>
<a href=/>Annulla</a>
  </body>
</html>
`

const mioArea1 = `
<html>
  <head>
<link rel="icon" href="/favicon.ico" />
<title>Votametro&trade;</title>
   <style type="text/css">
   .alrt { background-color: red }
   .wrn { background-color: yellow }
   .cen { margin-left:auto; margin-right:auto; }
   table
   {
   border-collapse:collapse;
   }
   table,th, td
   {
   border: 1px solid black;
   }
   </style>
  </head>
  <body>
  <h1>Votametro!</h1>
  <h2>Hello <i>%s</i></h2>
  <h3>Crea una nuova Area di voto</h3>
  <form action=/newdo method=post>
  <table>
  <tr>
   <th>Nome</th><th>Descrizione</th>
  </tr>
  <tr>
   <td><input type=text name=nm></td>
   <td><input type=text name=de></td>
  </tr>
  </table>
  <blockquote>
  <table>
`

const mioArea2 = `
  <tr><td>Voce %d&nbsp;</td><td><input type=text name=v%d></td></tr>
`

const mioArea3 = `
  </table>
 <p>Numero massimo di voci: %d
  </blockquote>
  <button type=submit>Conferma</button>
  </form>
  <a href=/miovota>Annulla</a>
  </body>
</html>
`

