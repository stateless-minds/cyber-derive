function setupCompetition(waitingForPeers) {

  if (waitingForPeers) {
    startTimeout();
  } else {
    interruptTimeout();
  }
}

let timeoutId;

function startTimeout() {
  const overlay = document.querySelector('.overlay');
  const textElement = document.getElementById('countdown');
  const waitingText = [
    "Waiting for players to join..."
  ];

  let index = 0;

  let intervalDelay = 10000; // initial 10 seconds delay
  let hideOverlayDelay = 10000; // delay for hiding overlay

  timeoutId = setTimeout(() => {
    overlay.style.display = 'flex';
    overlay.style.color = 'gold';
    textElement.textContent = waitingText[index];
    textElement.style.opacity = '1';

    // Recursive function to update text with dynamic delay
    function updateText() {
      index++;
      if (index < waitingText.length) {
        textElement.textContent = waitingText[index];

        // Dynamically set next delay here (can be updated before calling updateText)
        setTimeout(updateText, intervalDelay);
      } else {
        textElement.style.opacity = '0';

        // Set timeout to hide overlay with dynamic delay
        setTimeout(() => {
          overlay.style.display = 'none';
        }, hideOverlayDelay);
      }
    }
    
    updateText();
  }, 0); // no delay
}

function interruptTimeout() {
  if (timeoutId) {
    // clearTimeout(timeoutId);
    interrupted = true;

    calculateWinner();  // Highlight winner avatar after overlay disappears
  }
}

function getRandomSmall(epsilon) {
  return Math.random() * epsilon;
}

function calculateWinner() {
    const stats = document.querySelectorAll('.stats-container');
    if (stats.length === 0) return;

    let maxTotal = 0;
    let maxStat = null; // to store the stat element with max total

    for (const stat of stats) {
      const energy = Number(stat.querySelector('#energy').textContent);
      const speed = Number(stat.querySelector('#speed').textContent);
      const safety = Number(stat.querySelector('#safety').textContent);

      const speedMultiplier = 1.1
      const safetyMultiplier = 1.2
      const epsilon = 0.0001

      const total = energy + speed * speedMultiplier + safety * safetyMultiplier + getRandomSmall(epsilon);

      if (total > maxTotal) {
        maxTotal = total;
        maxStat = stat;
      }
    }

    const compItemParent = maxStat.closest('.comp-item');

    const updateDelivery = new CustomEvent('delivery-updated', {
      detail: {
          courierId: JSON.stringify(compItemParent.id),
      }
    });
  
    var elementMap = document.getElementById("comp")
  
    elementMap.dispatchEvent(updateDelivery);


    selectWinner(compItemParent);
    
    // Add highlight style and blinking animation
    compItemParent.classList.add('highlight', 'blink-green');
    
    // Remove blinking after 3 seconds but keep highlight
    setTimeout(() => {
      compItemParent.classList.remove('blink-green');
    }, 3000);
  }

  function selectWinner(playerDiv) {
    // Scroll the div into view smoothly
    playerDiv.scrollIntoView({ behavior: 'smooth', block: 'center' });

    // Create overlay element
    const overlay = document.createElement('div');
    overlay.className = 'winner-overlay';
    overlay.textContent = 'Winner';
  
    // Append overlay to player div
    playerDiv.appendChild(overlay);
  }