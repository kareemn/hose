
function output(str) {
    console.log(str);
}

// Events
// init() once the page has finished loading.
window.onload = init;

var context;
var source;
var analyser;
var buffer;
var audioBuffer;

var analyserView1;

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


function Float32Concat(first, second)
{
    var firstLength = first.length,
        result = new Float32Array(firstLength + second.length);

    result.set(first);
    result.set(second, firstLength);

    return result;
}

var bufferSize = 4096 * 1;
var decodedAudioBuffer = new Float32Array();

function onReceivePcm(message) {
    decodedAudioBuffer = Float32Concat(decodedAudioBuffer, new Float32Array(message.data));
//	console.log(ba);
}

function gotStream(stream) {
	//var client = new BinaryClient('ws://localhost:4000/audio/kimo');
	var audiows = new WebSocket("ws://localhost:4000/audio/kimo");
	audiows.binaryType = "arraybuffer";
	audiows.onopen = function() {
       // Web Socket is connected, send data using send()
       // ws.send("Message to send");
       console.log("Message is sent...");
       window.Stream = audiows;
    };
    audiows.onmessage = function (pcmdata) {
		   onReceivePcm(pcmdata);
	}
	/*
	client.on('open', function() {
	  // for the sake of this example let's put the stream in the window
	  window.Stream = client.createStream();
	  window.Stream.on('data', function(data) {

	   });
	});
	*/
    initAudio(stream);
}

function init() {
    getUserMedia({audio:true}, gotStream);
}

function convertFloat32ToInt16(buffer) {
  l = buffer.length;
  buf = new Int16Array(l);
  while (l--) {
    buf[l] = Math.min(1, buffer[l])*0x7FFF;
  }
  return buf.buffer;
}

function recorderProcess(e) {
  var left = e.inputBuffer.getChannelData(0);
  window.Stream.send(left);
}

var audioContext = new window.AudioContext();
var audioNode = audioContext.createScriptProcessor(4096, 1, 1);

function initPlayback() {

	audioNode.onaudioprocess = function(event) {
		console.log("audioNode");
        if (decodedAudioBuffer.length >= bufferSize) {
	           if (decodedAudioBuffer === undefined) {
		  	       return;
	           }
               var decoded = decodedAudioBuffer;
               decodedAudioBuffer = new Float32Array();
               var output = event.outputBuffer.getChannelData(0);
                 for (var i = 0; i < output.length; i++) {
                   output[i] = decoded[i];
                 }
        }
	};
	audioNode.connect(audioContext.destination);
}

function initAudio(stream) {
	initPlayback();

	var audioInput = audioContext.createMediaStreamSource(stream);
	var bufferSize = 2048;
	// create a javascript node
	var recorder = audioContext.createScriptProcessor(bufferSize, 1, 1);
	// specify the processing function
	recorder.onaudioprocess = recorderProcess;
	// connect stream to our recorder
	audioInput.connect(recorder);
	// connect our recorder to the previous destination
	recorder.connect(audioContext.destination);
	
}

