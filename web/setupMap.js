function setupMap(peerId, action, deliveriesJSON) {
  if ("geolocation" in navigator) {
    navigator.geolocation.getCurrentPosition(
      (position) => {
        let lat = position.coords.latitude
        let long = position.coords.longitude

        let map = L.map('map', {
          minZoom: 13,
          maxZoom: 20,
        });
      
        if (window.location.hash) {
          let parts = window.location.hash.substring(1).split('/');
          map.setView([parts[2], parts[1]], parts[0]);
        } else {
          map.setView([lat, long], 13);
        }

        L.maplibreGL({
          style: 'https://tiles.openfreemap.org/styles/bright',
        }).addTo(map)

        // 1. Create a custom pane on your map
        map.createPane('tooltipPaneIntro');

        // 2. Set the z-index so it appears above default layers (e.g., above markers but below popups)
        map.getPane('tooltipPaneIntro').style.zIndex = 650;

        // 3. Enable pointer events on this pane so it can capture clicks
        map.getPane('tooltipPaneIntro').style.pointerEvents = 'auto';

        if (action == "request") {
          var tooltip = L.tooltip([lat, long], {content: `<center><h1>Welcome to Cyber Dérive!</h1></center>
            <div class="tooltip-content">
              <h2>Request a delivery: </h2>
              <ul>
                <li>Choose a pickup and dropoff location</li>
                <li>Select delivery date and time</li>
                <li>Select date and time for confirmation</li>
                <li>Confirm delivery was completed</li>
              </ul>
            </div>`, direction: 'top', permanent: true, interactive: true, pane: 'tooltipPaneIntro',  // <-- important: assign to custom pane
            className: 'intro-tooltip'}).addTo(map);

            tooltip.on('click', function(e) {
              tooltip.remove();
              L.DomEvent.stopPropagation(e);
            });
        } else if (action == "deliver") {
          var tooltip = L.tooltip([lat, long], {content: `<center><h1>Welcome to Cyber Dérive!</h1></center>
            <div class="tooltip-content">
              <h2>Make a delivery: </h2>
              <ul>
                <li>Select a delivery request</li>
                <li>Equip your avatar with items</li>
                <li>Win the competition</li>
                <li>Await confirmation of your delivery</li>
                <li>Get new cards and improve your avatar</li>
              </ul>
              <p><b>Important:</b></p>
              <p>Make sure to come back precisely at 'Enter by' time for the competition.
            </div>`, direction: 'top', permanent: true, interactive: true, pane: 'tooltipPaneIntro',  // <-- important: assign to custom pane
            className: 'intro-tooltip'}).addTo(map);

            tooltip.on('click', function(e) {
              tooltip.remove();
              L.DomEvent.stopPropagation(e);
            });
        }
        
        let counter = 0

        if (deliveriesJSON.length > 0) {
          // Parse the main JSON string
          const parsedData = JSON.parse(deliveriesJSON);

          var deliveries

          // Parse nested stringified JSON fields
          deliveries = parsedData.map(item => ({
            _id: item._id,
            coordinates: item.coordinates,
            delivery_datetime: item.delivery_datetime,
            entry_datetime: item.entry_datetime,
            created_by: item.created_by,
            participants: item.participants,
            courier_id: item.courier_id,
            status: item.status,
          }));

          for (const delivery of deliveries) {
            var markerPickup = L.marker([delivery.coordinates[0][0].toString(), delivery.coordinates[0][1].toString()]).addTo(map);
            markerPickup.bindPopup("<p><center><b>Delivery ID: "+delivery._id+"</b></center></p>", { "closeButton": false});

            var statusMsg = `<p>Status: `+delivery.status+`</p>`
            var courierMsg = `<p style="word-wrap: break-word;">To be delivered by: `+delivery.courier_id+`</p>`
            var deliveredMsg = `<div style="display: flex; align-items: center; gap: 20px; margin-left: 20%">
            <button onclick="deliveryCompleted('${delivery._id}')">Delivered</button>
            <button onclick="deliveryFailed('${delivery._id}')">Not Delivered</button>
            </div>`
            var makeDelivery = `<button onclick="redirectDelivery('${delivery._id}')">Make Delivery</button>`
            var button

            if (action == "request") {
              button = statusMsg

              if (delivery.status == "ready_to_deliver") {
                button = button + courierMsg + deliveredMsg
              }
            } else {
              if (delivery.participants.includes(peerId) || delivery.status != "pending")  {
                button = statusMsg
              } else if (delivery.status == "pending") {
                button = makeDelivery
              }
            }

            var markerDropoff = L.marker([delivery.coordinates[1][0].toString(), delivery.coordinates[1][1].toString()]).addTo(map);
            markerDropoff.bindPopup(`<center><b>Delivery ID: `+delivery._id+`
              <br><br>
              <label for="entry-date">Enter by:</label>
                <input id="entry-datetime" type='datetime-local' value=`+delivery.entry_datetime+` disabled />
              <br><br>
              <label for="delivery-date">Deliver by:</label>
                <input id="delivery-datetime" type='datetime-local' value=`+delivery.delivery_datetime+` disabled />
              <br><br>
              `+button+`
              </center></b>`, { "closeButton": false, "className": "comp"});
            L.polyline(delivery.coordinates, {color: 'green'}).addTo(map);
          }
        }

        var pickupLat, pickupLng, dropoffLat, dropoffLng

        function onMapClick(e) {
          switch (counter) {
            case 0:
              var marker = L.marker([e.latlng.lat.toString(), e.latlng.lng.toString()]).addTo(map);
              
              marker.on('click', function(e) {
                marker.remove();
                polyline.remove()
              });

              marker.bindPopup(`<p><center><b>
                  Pickup
                </b></center></p>
                <i>Wrong click? Click the marker again to remove it.</i>`, { "closeButton": false}).openPopup();
              pickupLat = e.latlng.lat.toString()
              pickupLng = e.latlng.lng.toString()
              counter++
              return
            case 1:
              var marker = L.marker([e.latlng.lat.toString(), e.latlng.lng.toString()]).addTo(map);

              marker.on('click', function(e) {
                marker.remove();
                polyline.remove()
              });

              dropoffLat = e.latlng.lat.toString()
              dropoffLng = e.latlng.lng.toString()
              var latlngs = [
                [pickupLat, pickupLng],
                [dropoffLat, dropoffLng]
              ]

              const now = new Date();

              let plusOneHour = new Date(now.getTime());
              let plusTwoHours = new Date(now.getTime());

              let xMins = new Date();

              // Add one hour to current time
              plusOneHour.setHours(plusOneHour.getHours() + 1);

              // Add 5 minutes to current time testing
              xMins.setMinutes(now.getMinutes() + 5);

              // Add two hours to current time
              plusTwoHours.setHours(plusTwoHours.getHours() + 2);

              const entryDateTimeValue = getFormattedDateTime(xMins);
              const deliveryDateTimeValue = getFormattedDateTime(plusTwoHours);

              marker.bindPopup(`<p><center><b>Dropoff
                <br><br>
                <label for="entry-datetime">Enter by:</label>
                 <input id="entry-datetime" type='datetime-local' value='${entryDateTimeValue}' required />
                <br><br>
                <label for="delivery-datetime">Deliver by:</label>
                  <input id="delivery-datetime" type='datetime-local' value='${deliveryDateTimeValue}' required />
                <br><br>
                `+createButton(action, latlngs)+`
                </p></center></b>
                <i>Wrong click? Click the marker again to remove it.</i>`, { "closeButton": false, "className": "comp"}).openPopup();
                
              var polyline = L.polyline(latlngs, {color: 'red'}).addTo(map);

              validateDateTime();

              pickupLat = 0
              pickupLng = 0
              dropoffLat = 0
              dropoffLng = 0

              counter++
              return
            default: 
              counter = 0
              var marker = L.marker([e.latlng.lat.toString(), e.latlng.lng.toString()]).addTo(map);
              
              marker.on('click', function(e) {
                marker.remove();
                polyline.remove()
              });

              marker.bindPopup(`<p><center><b>
                  Pickup
                </b></center></p>
                <i>Wrong click? Click the marker again to remove it.</i>`, { "closeButton": false}).openPopup();
              pickupLat = e.latlng.lat.toString()
              pickupLng = e.latlng.lng.toString()
              counter++
              return
          }
        }

        if (action == "request") {
          map.on('click', onMapClick);
        }
      },
      (error) => {
        alert("In order to use Cyber Dérive location needs to be enabled");
      }
    );
  } else {
    console.error("Geolocation is not supported by this browser.");
  }
}

function createButton(action, latlngs) {
  if (action == "request") {
    // Serialize latlngs as JSON string to pass in inline handler
    const latlngsStr = JSON.stringify(latlngs);
    return `<button onclick='requestDelivery(${latlngsStr})'>Request Delivery</button>`;
  }
}

function requestDelivery(latlngs) {
  deliveryID = crypto.randomUUID()

  const inputs = [
  document.getElementById("delivery-datetime"),
  document.getElementById("entry-datetime"),
  ];

  for (const input of inputs) {
    if (!input.checkValidity()) {
      input.reportValidity();
      return;
    }
  }

  const createDelivery = new CustomEvent('delivery-created', {
    detail: {
        id: JSON.stringify(deliveryID),
        latlngs: JSON.stringify(latlngs),
        deliveryDateTime: JSON.stringify(inputs[0].value),
        entryDateTime: JSON.stringify(inputs[1].value),
    }
  });

  var elementMap = document.getElementById("map")

  elementMap.dispatchEvent(createDelivery);
}

function removeMarker(marker) {
  marker.remove();
}

if (window.location.hash === "#close") {
  window.location.replace("/map");
}

function validateDateTimes() {
  const enterByDateTime = document.getElementById('entry-datetime');
  const deliverByDateTime = document.getElementById('delivery-datetime');

  let enterDateTime = new Date(enterByDateTime.value);
  let deliverDateTime = new Date(deliverByDateTime.value);

  const now = new Date();

  if (enterByDateTime.value && deliverByDateTime.value) {
    // oneHour = isAtLeastXHoursAfter(now.getTime(), enterDateTime.getTime(), 1);
    
    // if (!oneHour) {
    //   alert("'Enter By' time must be at least an hour from now.");
    //   // Optionally reset the invalid input or prevent form submission here
    //   enterByDateTime.value = "";
    //   return;
    // }

    xMins = isAtLeastXMinutesAfter(now.getTime(), enterDateTime.getTime(), 5);
    if (!xMins) {
      alert("'Enter By' time must be at least 5 minutes from now.");
      // Optionally reset the invalid input or prevent form submission here testing
      enterByDateTime.value = "";
      return;
    }

    twoHours = isAtLeastXHoursAfter(now.getTime(), deliverDateTime.getTime(), 2);

    if (!twoHours) {
      alert("'Deliver By' time must be at least two hours from now.");
      // Optionally reset the invalid input or prevent form submission here
      deliverByDateTime.value = "";
      return;
    }

    if (enterDateTime.getTime() === deliverDateTime.getTime()) {
      alert("'Enter By' datetime can not be same as 'Deliver By' datetime.");
      // Optionally reset the invalid input or prevent form submission here
      enterByDateTime.value = "";
      deliverByDateTime.value = "";
      return;
    } else if (enterDateTime.getTime() > deliverDateTime.getTime()) {
      alert("'Enter By' datetime can not be after 'Deliver By' datetime.");
      // Optionally reset the invalid input or prevent form submission here
      enterByDateTime.value = "";
      deliverByDateTime.value = "";
      return;
    } else if (!isAtLeastXHoursAfter(enterDateTime.getTime(), deliverDateTime.getTime(), 1)) {
      alert("'Enter By' time must be at least an hour before 'Deliver By' time.");
      // Optionally reset the invalid input or prevent form submission here
      enterByDateTime.value = "";
      deliverByDateTime.value = "";
      return;
    }
  }
}

function isAtLeastXMinutesAfter(fromDateTime, toDateTime, minutes) {
  // Convert datetime-local string to Date object
  const toDate = new Date(toDateTime);
  const xMinutesFromNow = new Date(fromDateTime +  minutes * 60 * 1000);
  
  // Check if inputDate is at least x minutes from now
  return toDate >= xMinutesFromNow;
}

function isAtLeastXHoursAfter(fromDateTime, toDateTime, hours) {
  // Convert datetime-local string to Date object
  const toDate = new Date(toDateTime);
  const xHoursFromNow = new Date(fromDateTime + (hours * 60 * 60 * 1000));
  
  // Check if inputDate is at least x hours from now
  return toDate >= xHoursFromNow;
}

function validateDateTime() {
  const enterByDateTime = document.getElementById('entry-datetime');
  const deliverByDateTime = document.getElementById('delivery-datetime');

  enterByDateTime.addEventListener('change', validateDateTimes);
  deliverByDateTime.addEventListener('change', validateDateTimes);
}

function validateTimes() {
  const enterByTime = document.getElementById('entry-time');
  const deliverByTime = document.getElementById('delivery-time');

  // Parse 'HH:MM' time string into hour and minute numbers
  const [enterHour, enterMinute] = enterByTime.value.split(':').map(Number);
  const [deliverHour, deliverMinute] = deliverByTime.value.split(':').map(Number);

  const now = new Date(); // current date and time

  let enterTime = new Date(now.getFullYear(), now.getMonth(), now.getDate(), enterHour, enterMinute);
  let deliverTime = new Date(now.getFullYear(), now.getMonth(), now.getDate(), deliverHour, deliverMinute);
  let deliverTimeMinusOneHour = new Date(now.getFullYear(), now.getMonth(), now.getDate(), deliverHour-1, deliverMinute);

  let timeNowplusOneHour = new Date(now.getFullYear(), now.getMonth(), now.getDate(), now.getHours()+1, now.getMinutes());
  let timeNowplusTwoHours = new Date(now.getFullYear(), now.getMonth(), now.getDate(), now.getHours()+2, now.getMinutes());



  if (enterByTime.value && deliverByTime.value) {
    if (enterTime < timeNowplusOneHour || deliverTime < timeNowplusTwoHours) {
      alert("'Enter By' time must be at least an hour from now and 'Deliver By' time 2 hours from now.");
        // Optionally reset the invalid input or prevent form submission here
        enterByTime.value = "";
        deliverByTime.value = "";
    } else if (enterTime > deliverTimeMinusOneHour) {
        alert("'Enter By' time must be at least an hour before 'Deliver By' time.");
        // Optionally reset the invalid input or prevent form submission here
        enterByTime.value = "";
        deliverByTime.value = "";
    }
  }
}

function validateTime() {
  const enterByTime = document.getElementById('entry-time');
  const deliverByTime = document.getElementById('delivery-time');

  enterByTime.addEventListener('change', validateTimes);
  deliverByTime.addEventListener('change', validateTimes);
}

// Helper to pad single digit numbers with a leading zero
const pad = (num) => num.toString().padStart(2, '0');

// DateTime formatter helper
const getFormattedDateTime = (date) => {
  const year = date.getFullYear();
  const month = pad(date.getMonth() + 1);
  const day = pad(date.getDate());
  const hours = pad(date.getHours());
  const minutes = pad(date.getMinutes());
  return `${year}-${month}-${day}T${hours}:${minutes}`;
};

function deliveryCompleted(id) {
  const completeDelivery = new CustomEvent('delivery-completed', {
    detail: {
        id: JSON.stringify(id),
    }
  });

  var elementMap = document.getElementById("map")

  elementMap.dispatchEvent(completeDelivery);
}

function deliveryFailed(id) {
  const failedDelivery = new CustomEvent('delivery-failed', {
    detail: {
        id: JSON.stringify(id),
    }
  });

  var elementMap = document.getElementById("map")

  elementMap.dispatchEvent(failedDelivery);
}

function redirectDelivery(id) {
  const redirectDeliveryEvent = new CustomEvent('delivery-redirected', {
    detail: {
        id: JSON.stringify(id),
    }
  });

  console.log(redirectDeliveryEvent);

  var elementMap = document.getElementById("map")

  elementMap.dispatchEvent(redirectDeliveryEvent);
}