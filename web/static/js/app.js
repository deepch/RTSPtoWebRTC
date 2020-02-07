let config = {
    iceServers: [ {urls: 'stun:stun.l.google.com:19302'  }]
  };
const pc = new RTCPeerConnection(config);
console.log(pc)
function isSafari() {
    return browser() === 'safari';
  }
  function browser() {
  const ua = window.navigator.userAgent.toLocaleLowerCase();

  if (ua.indexOf('edge') !== -1) {
    return 'edge';
  } else if (ua.indexOf('chrome') !== -1 && ua.indexOf('edge') === -1) {
    return 'chrome';
  } else if (ua.indexOf('safari') !== -1 && ua.indexOf('chrome') === -1) {
    return 'safari';
  } else if (ua.indexOf('opera') !== -1) {
    return 'opera';
  } else if (ua.indexOf('firefox') !== -1) {
    return 'firefox';
  }
  return;
}

let log = msg => {
  document.getElementById('div').innerHTML += msg + '<br>'
}

pc.ontrack = function (event) {
  console.log("ontrack")
  var el = document.createElement(event.track.kind)
  el.srcObject = event.streams[0]
  el.muted    = true
  el.autoplay = true
  el.controls = true
  el.width    = 600
  document.getElementById('remoteVideos').appendChild(el)
}

pc.oniceconnectionstatechange = e => log(pc.iceConnectionState)
pc.onicecandidate = event => {
   console.log("onicecandidate")
  if (event.candidate === null) {
    document.getElementById('localSessionDescription').value = btoa(pc.localDescription.sdp)
    var suuid = $('#suuid').val();
    $.post("/recive", { suuid: suuid,data:btoa(pc.localDescription.sdp)} ,function(data){
      document.getElementById('remoteSessionDescription').value = data
      window.startSession()
    });
  }
}

pc.createOffer({offerToReceiveVideo: true, offerToReceiveAudio: false}).then(d => pc.setLocalDescription(d)).catch(log)

window.startSession = () => {
  let sd = document.getElementById('remoteSessionDescription').value
  if (sd === '') {
    return alert('Session Description must not be empty')
  }
  try {
    pc.setRemoteDescription(new RTCSessionDescription({type: 'answer', sdp: atob(sd)}))
  } catch (e) {
    alert(e)
  }
}

$(document).ready(function() {
  var suuid = $('#suuid').val();
  $('#'+suuid).addClass('active');
});
