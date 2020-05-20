require('./bootstrap');
import Masonry from 'masonry-layout'
import ImagesLoaded from 'imagesloaded'
import InfiniteLoading from 'vue-infinite-loading'

window.Vue = require('vue');
Vue.options.delimiters = ["${", "}"];

window._base64ToArrayBuffer = function (base64) {
	base64 = base64.replace(/^data:image\/(png|jpeg);base64,/, '')
	var binaryString = window.atob(base64)
	var len = binaryString.length
	var bytes = new Uint8Array(len)
	for (var i = 0; i < len; i++) {
		bytes[i] = binaryString.charCodeAt(i)
	}
	return bytes.buffer
}

import {
	fetchPost,
	fetchPostStatus
} from './ajax'

document.addEventListener('DOMContentLoaded', () => {
	let isLoading

	$(document).on('click', '.aclike', function (e) {
		e.preventDefault()
		let elemCount = $(this).parent().find('.count')
		let oldNumber = isNaN(elemCount.text())
		var postID = $(this).attr('data-post-id')
		if (isLoading) return
		if (!isSignin) {
			window.location.href = `/signin?r=/post/${postID}`
			return
		}
		fetchPost('/ajax/like', {
				id: postID + ""
			})
			.then(res => {
				if (res.isLike) {
					if (!oldNumber) {
						elemCount.text(parseInt(elemCount.text()) + 1)
					}
					$(this).find('i').removeClass('fal _cl-gray-9').addClass('fas _cl-yellow-1')
				} else {
					if (!oldNumber) {
						elemCount.text(parseInt(elemCount.text()) - 1)
					}
					$(this).find('i').removeClass('fas _cl-yellow-1').addClass('fal _cl-gray-9')
				}
			})
	})

	$(document).on('click', '.acfollow', function (e) {
		e.preventDefault()
		var gapID = $(this).attr('data-gap-id')
		if (isLoading) return
		isLoading = true
		if (!isSignin) {
			window.location.href = `/signin?r=/post/${gapID}`
			return
		}
		fetchPost('/ajax/follow/gap', {
				id: gapID + ""
			})
			.then(res => {
				if (res.isfollow) {
					$(this).find('i').removeClass('fal fa-plus').addClass('fal fa-minus')
					$(this).find('.followtext').text('เลิกติดตาม')

				} else {
					$(this).find('i').removeClass('fal fa-minus').addClass('fal fa-plus')
					$(this).find('.followtext').text('ติดตาม')
				}
				isLoading = false
			}).catch(error => {
				isLoading = false
			})
	})

	$(document).on('click', '.acfollowtopic', function (e) {
		e.preventDefault()
		var topicID = $(this).attr('data-topic-id')
		if (isLoading) return
		isLoading = true
		fetchPost('/ajax/follow/topic', {
				id: topicID + ""
			})
			.then(res => {
				if (res.isfollow) {
					$(this).text('เลิกติดตาม')
				} else {
					$(this).text('ติดตาม')
				}
				isLoading = false
			}).catch(error => {
				isLoading = false
			})
	})


	//true = หัว
	//false = ก้อย

	$(document).on('click', '.20-coin', function (e) {
		e.preventDefault()

		fetchPost('/ajax/frontback/bet', {
			frontback: true,
			price: 20
			})
			.then(res => {
				console.log(res)
				$('.wallet').text(res.wallet)
			}).catch(error => {

			})
	})


	// $(document).ready(function() {
	// 	$.get("https://api.tradingeconomics.com/markets/search/Crude%20Oil?c=guest:guest", function(data) {
	// 		// console.log(data[0].Last)
	// 		// var value = parseFloat(data[0].Last)
	// 		$(".number-oil").text(parseFloat(data[0].Last).toFixed(2))
	// 		$(".number-dailychange").text(" +" + parseFloat(data[0].DailyChange).toFixed(2) + "%")
	// 		$(".number-dailypercentualchange").text(" +" + parseFloat(data[0].DailyPercentualChange).toFixed(2) + "%")
	// 		// $(".number-oil").text($(".number-oil").val(parseFloat(data[0].Last).toFixed(2)));

	// 	});
	// 	setInterval(getoil, 5000);
	// });


	// function getoil() {
	// 	$.get( "https://api.tradingeconomics.com/markets/search/Crude%20Oil?c=guest:guest", function(data) {
	// 		// console.log(data[0].Last)
	// 		// var value = parseFloat(data[0].Last)
	// 		$(".number-oil").text(parseFloat(data[0].Last).toFixed(2))
	// 		$(".number-dailychange").text(" +" + parseFloat(data[0].DailyChange).toFixed(2) + "%")
	// 		$(".number-dailypercentualchange").text(" +" + parseFloat(data[0].DailyPercentualChange).toFixed(2) + "%")
	// 	});
	//   }



	//https://api.tradingeconomics.com/markets/search/Crude%20Oil?c=guest:guest




	$('.acsearch').on('click', function (e) {
		e.preventDefault()
		window.location.href = "/search"
	})
	$('.acback').on('click', function (e) {
		e.preventDefault()
		window.history.back()
	})
	jQuery(function ($) {
		$(".sidebar-dropdown > a").click(function () {
			$(".sidebar-submenu").slideUp(200)
			if ($(this).parent().hasClass("active")) {
				$(".sidebar-dropdown").removeClass("active")
				$(this).parent().removeClass("active")
			} else {
				$(".sidebar-dropdown").removeClass("active")
				$(this).next(".sidebar-submenu").slideDown(200)
				$(this).parent().addClass("active")
			}
		})
	})
}) // end //

// let postInstant = new Vue({
// 	data() {
// 		return {
// 			postList: [],
// 			next: new Date(0),
// 			type: '',
// 			masonry: null,
// 			isAjax: false,
// 		}
// 	},

// 	components: {
// 		InfiniteLoading
// 	},

// 	computed: {
// 		hasPost() {
// 			return this.postList.length > 0
// 		}
// 	},

// 	methods: {
// 		infiniteHandler($state) {
// 			fetchPostStatus(postNextURL, {
// 				type: this.type,
// 				next: this.next
// 			}).then((res) => {
// 				if (res.status === 200) {
// 					this.isAjax = true
// 					this.next = res.data.next
// 					this.postList = this.postList.concat(res.data.post)
// 					this.initMsy($state)
// 				} else if (res.status === 204) {
// 					$state.complete()
// 				}
// 			}).catch(() => {
// 				$state.complete()
// 			})
// 		},

// 		initMsy($state) {
// 			this.$nextTick(() => {
// 				const container = this.$refs.postList
// 				this.masonry = new Masonry(container, {
// 					itemSelector: '.post-item',
// 					percentPosition: true,
// 					transitionDuration: '0.7s'
// 				})
// 				ImagesLoaded(container).on('progress', function () {
// 					this.masonry.layout()
// 				}.bind(this))
// 				$state.loaded()
// 			})
// 		},

// 		selectType(value) {
// 			this.type = value
// 			this.postList = []
// 			this.next = new Date(0)
// 			this.$nextTick(() => {
// 				if (this.masonry) {
// 					this.masonry.destroy()
// 				}
// 				this.$refs.infiniteLoading.$emit('$InfiniteLoading:reset')
// 			})
// 		}

// 	}
// })
 // end post instant



let pananInstant = new Vue({

	data() {
		return {
			coin: 0,
			postBetURL: '/ajax/frontback/bet',
			isCoin: 0
		}
	},

	components: {
	},

	computed: {

	},

	methods: {
		changeCoin(c, type, finalfrontback) {
			if (type === "coin" && c > 0) {

				this.coin = c
				this.isCoin = c
				console.log("bet coin =" +this.coin  +finalfrontback)

			} else if (type === "bet" && this.isCoin > 0 && this.coin > 0) {
				fetchPost('/ajax/frontback/bet', {
					frontback: finalfrontback,
					price: this.coin
					})
					.then(res => {
						console.log("start bet" + this.coin)
						$('.wallet').text(res.wallet)
						this.coin = 0
						this.isCoin = 0
					}).catch(error => {

					})
			}
		},

	}
})


// end post instant

// let postSearch = new Vue({

// 	el: '#search-overlay',
// 	data() {
// 		return {
// 		routeSearch: '/ajax/search/navigationbar',
// 		open: false,
// 		search: '',
// 		masonry: null,
// 		searchListTopic: null,
// 		noData: false,
// 		postList: []
// 		}
// 	},
// 	mounted() {
// 		this.$nextTick(() => {
// 			 this.$refs.search.focus()
// 		})
// 	},
// 	methods: {
// 		submitSearch() {
// 			axios.post(this.routeSearch, {
// 				text: this.search
// 			}, { withCredentials: true }).then(res => {
// 				if (res.status === 200) {
// 					this.searchList = null
// 					this.searchListTopic = res.data.topic
// 					this.postList = res.data.post
// 					this.$nextTick(() => {
// 						const container = this.$refs.postList
// 						this.masonry = new Masonry(container, {
// 							itemSelector: '.post-item',
// 							percentPosition: true,
// 							transitionDuration: '0.7s'
// 						})
// 						ImagesLoaded(container).on('progress', function () {
// 							this.masonry.layout()
// 						}.bind(this))
// 					})
// 					this.open = true
// 					this.noData = false
// 				} else if (res.status === 204) {
// 					this.searchListTopic = null
// 					this.postList = null
// 					this.noData = true
// 					this.open = true
// 				}
// 			}).catch(() => {
// 				this.noData = true
// 			})
// 		},
// 		searchChange: _.debounce(function (value) {
// 			if (this.search && this.search.trim()) {

// 				this.submitSearch(value)
// 			} else {
// 				this.noData = false
// 				this.open = false
// 			}
// 		}, 300),
// 	}
// })



const panan = document.getElementById('panan-container')
if(panan){
	pananInstant.$mount(panan)
}

// const postCtn = document.getElementById('post-container')
// if (postCtn) {
// 	postInstant.$mount(postCtn)
// }