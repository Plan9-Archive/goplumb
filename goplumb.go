

/*2:*/


//line goplumb.w:52

// Copyright (c) 2013 Alexander Sychev. All rights reserved.
//
// Redistribution and use in source and binary forms, with or without
// modification, are permitted provided that the following conditions are
// met:
//
//    * Redistributions of source code must retain the above copyright
// notice, this list of conditions and the following disclaimer.
//    * Redistributions in binary form must reproduce the above
// copyright notice, this list of conditions and the following disclaimer
// in the documentation and/or other materials provided with the
// distribution.
//    * The name of author may not be used to endorse or promote products derived from
// this software without specific prior written permission.
//
// THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS
// "AS IS" AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT
// LIMITED TO, THE IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR
// A PARTICULAR PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT
// OWNER OR CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL,
// SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT
// LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE,
// DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY
// THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
// (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE
// OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.

// Package goplumb provides interface to plumber - interprocess messaging from Plan 9.
package goplumb

import(


/*6:*/


//line goplumb.w:191

"code.google.com/p/goplan9/plan9"
"code.google.com/p/goplan9/plan9/client"
"os"



/*:6*/



/*10:*/


//line goplumb.w:212

"sync"



/*:10*/



/*15:*/


//line goplumb.w:277

"fmt"



/*:15*/



/*17:*/


//line goplumb.w:295

"strings"



/*:17*/



/*20:*/


//line goplumb.w:336

"errors"
"io"



/*:20*/



/*25:*/


//line goplumb.w:418

"bytes"
"strconv"



/*:25*/


//line goplumb.w:84

)

type(


/*4:*/


//line goplumb.w:171

//Message desribes a plumber message.
Message struct{
Src string
Dst string
Wdir string
Type string
Attr Attrs
Data[]byte
}



/*:4*/



/*5:*/


//line goplumb.w:183

//Attrs is a map of an attribute of a plumber message.
Attrs map[string]string




/*:5*/



/*7:*/


//line goplumb.w:196

Plumb struct{
f*client.Fid


/*30:*/


//line goplumb.w:548

ch chan*Message



/*:30*/


//line goplumb.w:199

}



/*:7*/


//line goplumb.w:88

)

var(


/*9:*/


//line goplumb.w:206

fsys*client.Fsys
sp*Plumb
rp*Plumb



/*:9*/


//line goplumb.w:92

)



/*:2*/



/*12:*/


//line goplumb.w:225

//Open opens a specified port with a specified omode.
func Open(port string,omode uint8)(*Plumb,error){


/*11:*/


//line goplumb.w:216

{
var err error
new(sync.Once).Do(func(){fsys,err= client.MountService("plumb")})
if err!=nil{
return nil,err
}
}


/*:11*/


//line goplumb.w:228

var p Plumb
var err error
if p.f,err= fsys.Open(port,omode);err!=nil{
return nil,err
}
return&p,nil
}



/*:12*/



/*14:*/


//line goplumb.w:252

//Send sends a message.
func(this*Plumb)Send(message*Message)error{
if this==nil||this.f==nil||message==nil{
return os.ErrInvalid
}
b:=Pack(message)
// a workaround: \.{plumber} can't receive a message with length more that 8192-plan9.IOHDRSIZE

for len(b)> 0{
c:=8192-plan9.IOHDRSIZE
if len(b)<c{
c= len(b)
}
c,err:=this.f.Write(b[:c])
if err!=nil{
return err
}
b= b[c:]
}
return nil
}



/*:14*/



/*16:*/


//line goplumb.w:281

//Pack packs a message to []byte.
func Pack(message*Message)[]byte{
s:=fmt.Sprintf("%s\n%s\n%s\n%s\n%s\n%d\n",
message.Src,message.Dst,
message.Wdir,message.Type,
PackAttr(message.Attr),
len(message.Data))
b:=append([]byte{},[]byte(s)...)
return append(b,message.Data...)
}



/*:16*/



/*18:*/


//line goplumb.w:299

//PackAttr packs attr to string. If an attribute value contains a white space,
//a quote or an equal sign the value will be quoted.
func PackAttr(attr Attrs)string{
var s string
first:=true
for n,v:=range attr{
if!first{
s+= " "
}
first= false
if strings.ContainsAny(v," '=\t"){
s+= fmt.Sprintf("%s='%s'",n,strings.Replace(v,"'","''",-1))
}else{
s+= fmt.Sprintf("%s=%s",n,v)
}
}
return s
}



/*:18*/



/*19:*/


//line goplumb.w:320

//SendText sends a text-only message; it assumes Type is text and Attr is empty.
func(this*Plumb)SendText(src string,dst string,wdir string,data string)error{
m:=&Message{
Src:src,
Dst:dst,
Wdir:wdir,
Type:"text",
Data:[]byte(data)}
return this.Send(m)
}



/*:19*/



/*21:*/


//line goplumb.w:341

//Recv returns a received message or an error.
func(this*Plumb)Recv()(*Message,error){
if this==nil||this.f==nil{
return nil,os.ErrInvalid
}
b:=make([]byte,8192)
n,err:=this.f.Read(b)
if err!=nil{
return nil,err
}
m,r:=UnpackPartial(b[:n])
if m!=nil{
return m,nil
}
if r==0{
return nil,errors.New("buffer too small")
}
if r> len(b)-n{
b1:=make([]byte,r+n)
copy(b1,b)
b= b1
}else{
b= b[:n+r]
}
_,err= io.ReadFull(this.f,b[n:])
if err!=nil{
return nil,err
}
m,r= UnpackPartial(b)
if m!=nil{
return m,nil
}
return nil,errors.New("buffer too small")
}



/*:21*/



/*24:*/


//line goplumb.w:409

//Unpack return unpacked message.
func Unpack(b[]byte)*Message{
m,_:=UnpackPartial(b)
return m
}



/*:24*/



/*26:*/


//line goplumb.w:423

//UnpackPartial helps to unpack messages splited in peaces.
//The first call to UnpackPartial for a given message must be sufficient to unpack
//the header; subsequent calls permit unpacking messages with long data sections.
//For each call, b contans the complete message received so far.
//If the message is complete, a pointer to the resulting message m will be returned,
//and a number of remainings bytes r will be set to 0.
//Otherwise m will be nil and r will be set to the number of bytes
//remaining to be read for this message
//to be complete (recall that the byte count is in the header).
//Those bytes should be read by the caller, placed at location b[r:],
//and the message unpacked again.
//If an error is encountered, m will be nil and r will be zero.
func UnpackPartial(b[]byte)(m*Message,r int){
bb:=bytes.Split(b,[]byte{'\n'})
if len(bb)<6{
return nil,0
}
m= &Message{
Src:string(bb[0]),
Dst:string(bb[1]),
Wdir:string(bb[2]),
Type:string(bb[3]),
Attr:UnpackAttr(string(bb[4]))}
n,err:=strconv.Atoi(string(bb[5]))
if err!=nil{
return nil,0
}
i:=0
for j:=0;j<6;j++{
i+= len(bb[j])+1
}
if r= n-(len(b)-i);r!=0{
return nil,r
}
m.Data= make([]byte,n)
copy(m.Data,b[i:i+n])
return m,0
}



/*:26*/



/*28:*/


//line goplumb.w:495

//UnpackAttr unpack the attributes from s
func UnpackAttr(s string)Attrs{
attrs:=make(Attrs)
for i:=0;i<len(s);{
var n,v string
for;i<len(s)&&s[i]!='=';i++{
n+= s[i:i+1]
}
i++
if i==len(s){
break
}
if s[i]=='\''{
i++
for;i<len(s);i++{
if s[i]=='\''{
if i+1==len(s){
break
}
if s[i+1]!='\''{
break
}
i++
}
v+= s[i:i+1]
}
i++
}else{
for;i<len(s)&&s[i]!=' ';i++{
v+= s[i:i+1]
}

}
i++
attrs[n]= v
}
return attrs
}



/*:28*/



/*29:*/


//line goplumb.w:536

//Close closes a plumbing connection.
func(this*Plumb)Close(){
if this!=nil&&this.f!=nil{
this.f.Close()
this.f= nil
}
}





/*:29*/



/*31:*/


//line goplumb.w:552

// MessageChannel returns a channel of *Message from which messages can be read or error.
// First call of MessageChannel starts a goroutine to read messages put them to the channel.
// Subsequent calls of EventChannel will return the same channel.
func(this*Plumb)MessageChannel()(<-chan*Message,error){
if this==nil||this.f==nil{
return nil,os.ErrInvalid
}
if this.ch!=nil{
return this.ch,nil
}
this.ch= make(chan*Message)
go func(ch chan<-*Message){
for m,err:=this.Recv();err==nil;m,err= this.Recv(){
ch<-m
}
close(ch)
}(this.ch)
return this.ch,nil
}



/*:31*/


