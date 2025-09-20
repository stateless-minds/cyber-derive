let swapped = false;

function setupAvatar() {
  // Variable to keep track of dragged card element
  let draggedCard = null;

  // Helper: Set appropriate ARIA attributes for dragging
  function setAriaDragging(card, dragging) {
    card.setAttribute('aria-grabbed', dragging ? 'true' : 'false');
  }

  function updateButtonState() {
    // Disable button if itemCount is 0, else enable
    saveButton.disabled = (equippedItems === 0);
  }

  let equippedItems = 0;
  const saveButton = document.getElementById("save-avatar")

  setTimeout(() => {
    equippedItems = saveButton.value;

    // Initialize draggable cards
    document.querySelectorAll('.card').forEach(card => {
      card.addEventListener('dragstart', (ev) => {
        draggedCard = ev.currentTarget;
        ev.dataTransfer.setData('text/plain', ev.currentTarget.id);
        setAriaDragging(draggedCard, true);
        // Hide dragged element after short delay for drag effect
        // setTimeout(() => {
        //   ev.currentTarget.style.display = 'none';
        // }, 0);
      });

      card.addEventListener('dragend', (ev) => {
        if (draggedCard) {
          draggedCard.style.display = 'block';
          setAriaDragging(draggedCard, false);
          draggedCard = null;
        }
      });
    });
  }, 1000);

  // Setup avatar slots as drop targets
  const slots = document.querySelectorAll('.slot');
  slots.forEach(slot => {
    slot.addEventListener('dragover', ev => {
      ev.preventDefault();
      slot.classList.add('dragover');
    });
    slot.addEventListener('dragenter', ev => {
      ev.preventDefault();
      slot.classList.add('dragover');
    });
    slot.addEventListener('dragleave', ev => {
      slot.classList.remove('dragover');
    });
    slot.addEventListener('drop', ev => {
      ev.preventDefault();
      slot.classList.remove('dragover');
      const cardId = ev.dataTransfer.getData('text/plain');
      const card = document.getElementById(cardId);
      saveButton.removeAttribute("disabled");
      equippedItems++;

      console.log(equippedItems);
      
      updateButtonState();
      if (!card) return;

      // Get categories
      const cardCategory = card.getAttribute('data-category');
      const allowed = slot.getAttribute('data-accept').split(',');
      const tooltip = document.getElementById('tooltip');

      // Only allow drop if card's category is accepted by slot
      if (!allowed.includes(cardCategory)) {
        tooltip.textContent = 'Cannot equip this item here, try ' + cardCategory + ' instead.';
        // Position tooltip near mouse
        tooltip.style.left = (ev.clientX + 10) + 'px';
        tooltip.style.top = (ev.clientY + 10) + 'px';
        tooltip.style.display = 'block';

        setTimeout(() => {
          tooltip.style.display = 'none';
        }, 3000); // Hide tooltip after 1.2 seconds
        // Optional: visually signal failure
        slot.style.borderColor = 'red';
        setTimeout(() => slot.style.borderColor = 'gold', 500);
        return;
      }

      // Swap cards if slot already occupied by a card (ignore label div)
      const existingCard = Array.from(slot.children).find(el => el.classList && el.classList.contains('card'));

      if (existingCard && !swapped) {
        swapped = true;
        const oldCardStat = existingCard.getAttribute("data-stat");
        const newCardStat = card.getAttribute("data-stat");
        const existingCardValue = parseInt(existingCard.getAttribute("data-value"), 10);
        const newCardValue = parseInt(card.getAttribute("data-value"), 10);
        const oldStatsElement = document.getElementById(oldCardStat);
        const newStatsElement = document.getElementById(newCardStat);

        if (oldCardStat === newCardStat) {
          let currentVal = parseInt(oldStatsElement.textContent, 10);
          // Remove old card's value and add new card's value
          currentVal -= existingCardValue;
          currentVal += newCardValue;
          oldStatsElement.textContent = currentVal;
        } else {
          // Different stats, update both accordingly
          let currentOldVal = parseInt(oldStatsElement.textContent, 10);
          currentOldVal -= existingCardValue;
          oldStatsElement.textContent = currentOldVal;

          let currentNewVal = parseInt(newStatsElement.textContent, 10);
          currentNewVal += newCardValue;
          newStatsElement.textContent = currentNewVal;
        }

        // Swap cards
        const oldParent = card.parentNode;
        slot.replaceChild(card, existingCard);
        oldParent.appendChild(existingCard);

        // Update data-equipped attributes
        card.setAttribute('data-equipped', 'true');
        existingCard.removeAttribute('data-equipped');
      } else if (existingCard && swapped) {
        tooltip.textContent = 'You can swap a card once per competition.';
        // Position tooltip near mouse
        tooltip.style.left = (ev.clientX + 10) + 'px';
        tooltip.style.top = (ev.clientY + 10) + 'px';
        tooltip.style.display = 'block';

        setTimeout(() => {
          tooltip.style.display = 'none';
        }, 3000); // Hide tooltip after 1.2 seconds
        // Optional: visually signal failure
        slot.style.borderColor = 'red';
        setTimeout(() => slot.style.borderColor = 'gold', 500);
        return;
      } else {
        // Placing a new card in an empty slot
        const cardStat = card.getAttribute("data-stat");
        const cardValue = parseInt(card.getAttribute("data-value"), 10);
        const statsElement = document.getElementById(cardStat);
        let currentVal = parseInt(statsElement.textContent, 10);
        currentVal += cardValue;
        statsElement.textContent = currentVal;
        slot.appendChild(card);

        // Mark card as equipped
        card.setAttribute('data-equipped', 'true');
      }
    });
  });

  // Inventory allows dropping cards back for unequip
  // const inventory = document.getElementById('inventory');
  // inventory.addEventListener('dragover', ev => {
  //   ev.preventDefault();
  // });
  // inventory.addEventListener('drop', ev => {
  //   ev.preventDefault();
  //   const cardId = ev.dataTransfer.getData('text/plain');
  //   const card = document.getElementById(cardId);
  //   if (card && card.getAttribute('data-equipped') === 'true') {
  //     // Only update stats if card was equipped
  //     const cardStat = card.getAttribute("data-stat");
  //     const cardValue = parseInt(card.getAttribute("data-value"), 10);
  //     const statsElement = document.getElementById(cardStat);
  //     let currentVal = parseInt(statsElement.textContent, 10);
  //     currentVal -= cardValue;
  //     statsElement.textContent = currentVal;

  //     card.removeAttribute('data-equipped');
  //     saveButton.removeAttribute("disabled");
  //     equippedItems--;
  //     updateButtonState();
  //     inventory.appendChild(card);
  //   } else if (card) {
  //     // Card was never equipped, just move it without updating stats
  //     inventory.appendChild(card);
  //   }
  // });
}