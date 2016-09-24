window.onload = init;

function randomstring(L){
    var s= '';
    var randomchar=function(){
      var n= Math.floor(Math.random()*62);
      if(n<10) return n; //1-10
      if(n<36) return String.fromCharCode(n+55); //A-Z
      return String.fromCharCode(n+61); //a-z
    }
    while(s.length< L) s+= randomchar();
    return s;
}

function hashCode(str) { // java String#hashCode
    var hash = 0;
    for (var i = 0; i < str.length; i++) {
       hash = str.charCodeAt(i) + ((hash << 5) - hash);
    }
    return hash;
} 

function intToRGB(i){
    var c = (i & 0x00FFFFFF)
        .toString(16)
        .toUpperCase();

    return "00000".substring(0, 6 - c.length) + c;
}

user = randomstring(5);

function MainSocketLoop() {
  if ("WebSocket" in window) {
     console.log("WebSocket is supported by your Browser!");

     // Let us open a web socket
     var ws = new WebSocket("ws://localhost:4000/socket/kimo");
     ws.onopen = function() {
        // Web Socket is connected, send data using send()
        // ws.send("Message to send");
        console.log("Message is sent...");
     };
     ws.onmessage = function (message) { 
        var payload = message.data;
        console.log("Message is received...");
        console.log(payload);
        object = JSON.parse(payload);
        $('#incoming_water').css({ 'color': '#' + intToRGB(hashCode(object["id"]))});
        $('#incoming_water').text(object["id"] + ": " + object["transcript"]);
     };

     ws.onclose = function() { 
        // websocket is closed.
        console.log("connection closed"); 
     };

   return ws
  } else {
     // The browser doesn't support WebSocket
     alert("WebSocket NOT supported by your Browser!");
  }
}

ws = MainSocketLoop();

function error() {
    console.log('Stream generation failed.');
}

function getUserMedia(dictionary, callback) {
    try {
        navigator.webkitGetUserMedia(dictionary, callback, error);
    } catch (e) {
        alert('webkitGetUserMedia threw exception :' + e);
    }
}

function init() {
    // getUserMedia({audio:true}, gotStream);
    getUserMedia({audio:true}, startRecognizing);
}

function startRecognizing() {
    var create_email = false;
    var final_transcript = '';
    var recognizing = false;
    var ignore_onend;
    var start_timestamp;
    if (!('webkitSpeechRecognition' in window)) {
      console.log("No webkitSpeechRecognition");
    } else {
      var recognition = new webkitSpeechRecognition();
      recognition.continuous = true;
      recognition.interimResults = true;
      recognition.onstart = function() {
        recognizing = true;
        console.log("on start recognizing");
      };
      recognition.onerror = function(event) {
         console.log("onerror recognizing");
      };
      recognition.onend = function() {
        recognizing = false;
        if (ignore_onend) {
          return;
        }
        console.log("onend recognizing");
      };
      recognition.onresult = function(event) {
        var interim_transcript = '';
        for (var i = event.resultIndex; i < event.results.length; ++i) {
          if (event.results[i].isFinal) {
            final_transcript += event.results[i][0].transcript;
          } else {
            interim_transcript += event.results[i][0].transcript;
          }
        }
        console.log(final_transcript);
        console.log(interim_transcript);
        console.log(JSON.stringify({'id': user, 'transcript': interim_transcript}));
        ws.send(JSON.stringify({'id': user, 'transcript': interim_transcript}));
      };
    }
    recognition.start();
}
