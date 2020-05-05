export default function () {
	// init click action
	$(document).on('click touchstart', function (event) {
		if (!$(event.target).closest('.dropdown-toggle').length) {
			dropdown()
		}
	})
	initDropdown()
}

function initDropdown() {
	const dropdownBtn = document.querySelectorAll('.sw-dropdown')
	for (let i = 0; i < dropdownBtn.length; i++) {
		dropdownBtn[i].addEventListener('click', dropdown)
	}
}

function dropdown() {
	var Selector = {
		MENU: '.dropdown-menu'
	}
	var Classname = {
		SHOW: 'show'
	}

	if (!this) {
		const list = document.querySelectorAll(Selector.MENU)
		for (let i = 0; i < list.length; i++) {
			if (hasClass(list[i], Classname.SHOW)) {
				list[i].classList.remove(Classname.SHOW)
			}
		}
		return
	}

	this._menu = this.querySelectorAll(Selector.MENU)[0]

	if (!hasClass(this._menu, Classname.SHOW)) {
		this._menu.classList.add(Classname.SHOW)
		return
	}
	this._menu.classList.remove(Classname.SHOW)
}

function hasClass(element, cls) {
	return (' ' + element.className + ' ').indexOf(' ' + cls + ' ') > -1
}