new Vue({
	el: '#app',
	data() {
		return {
			routeTag: "/ajax/search/tagtopic",
			routeDraftImage: "/ajax/post/draft/article",
			provinceData: config.province,
			provinceIDSelected: config.provinceSelectedID || 1,
			gapList: config.gapList,
			gapMain: config.mainGap,
			searchProvince: '',
			searchTag: '',
			placeholder: 'เพิ่มสถานที่',
			searchInput: false,
			resultText: config.provinceName || 1,
			current: 0,
			currentTag: 0,
			tagSelected: config.tagTopics || [],
			tagList: [],
			topicList: [],
			noMoreAdd: false,
			invalidText: false,
			isDraft: false,
			lastestSave: '',
			titlePost: '',
			vdourlLink: '',
			descriptionPost: config.postDes || '',
			isLoading: false,
			isUpload: false,
			config: {
				limitTag: 5
			}
		}
	},

	mounted() {
		this.$nextTick(() => {
			this.titlePost = this.$refs.titleValue.value
			this.vdourlLink = this.$refs.vdourlValue.value
			this.initEditor()
		})
	},

	computed: {
		filteredProvinceList() {
			this.current = 0
			return this.provinceData.filter(data => {
				return data.name.toLowerCase().includes(this.searchProvince.toLowerCase())
			})
		},
		calDropdownStyle() {
			return this.dropdownStyle
		},
		displayResult() {
			if (this.resultText) {
				return this.resultText
			}
			return this.placeholder
		},
		tagWarning() {
			if (this.noMoreAdd) return `ใส่สูงสุดได้ ${this.config.limitTag} แท็ก`
			return 'ไม่อนุญาตอักษรพิเศษ'
		},
		placeholderTag() {
			if (this.tagSelected.length == this.config.limitTag) return ''
			return 'ใส่แท็กสำหรับคอนเทนต์นี้'
		},
		hasTagList() {
			return this.tagList.length > 0
		},
		jsonTagSelected() {
			return JSON.stringify(this.tagSelected)
		},
		isDisable() {
			if (this.titlePost.trim() === "" && $(this.descriptionPost).text().trim() === "") {
				return true
			}
			return false
		}
	},

	methods: {
		submitForm() {
			if (this.titlePost.trim() === "" && $(this.descriptionPost).text().trim() === "") return
			this.descriptionPost = editor.getData()
			if (this.isDisable) return
			if (this.isDraft) return
			if (this.isUpload) return
			if (this.isLoading) return
			this.isLoading = true
			this.$refs.formPost.submit()
		},

		// === tag, province ===

		fetchPost(url, data) {
			return fetch(url, {
				credentials: 'include',
				method: 'POST',
				body: JSON.stringify(data)
			}).then((res) => {
				if (res.ok && res.status == 200) {
					return res.json()
				}
				throw Error(res.statusText)
			})
		},
		onToggle(e) {
			this.$nextTick(() => {
				if (this.searchInput) {
					if (e.target !== this.$refs.sws_input ||
						e.target !== this.$refs.sws_dropdown) {
						this.$refs.sws_input.blur()
					}
				}
			})
		},
		sws_button() {
			this.searchInput = true
			this.$nextTick(() => {
				this.$refs.sws_input.focus()
			})
		},
		onBlur() {
			this.onEscKey()
		},
		onEscKey() {
			this.searchProvince = ''
			this.searchInput = false
		},
		onEnterKey() {
			if (this.filteredProvinceList.length === 0) return
			const option = this.filteredProvinceList[this.current]
			this.selected(option)
		},
		selected(option) {
			this.provinceIDSelected = option.id
			this.resultText = option.name
			this.onEscKey()
		},
		onDownKey() {
			if (this.current + 1 > this.filteredProvinceList.length - 1) return
			this.current += 1
			if (this.current > 2) {
				this.$refs.sws_dropdown.scrollTop += 31
			}
		},
		onUpKey() {
			if (this.current === 0) return
			this.current -= 1

			if (this.current < (this.filteredProvinceList.length - 1) - 2) {
				this.$refs.sws_dropdown.scrollTop -= 31
			}
		},
		checkTags: _.debounce(function (value) {
			this.getTag(value)
		}, 300),
		getTag(text) {
			this.tagList = []
			if (!this.searchTag) return
			if (this.tagSelected.length >= this.config.limitTag) return
			this.fetchPost(this.routeTag, {
				text: text
			}).then((res) => {
				this.tagList = res
			}).catch((err) => {
				this.tagList = []
				return err
			})
		},
		onEscKeyTag() {
			this.searchTag = ''
			this.currentTag = ''
			this.tagList = []
			this.noMoreAdd = false
			this.invalidText = false
		},
		searchDupTag(option) {
			if (!this.tagSelected.length) return true
			let search = _.findIndex(this.tagSelected, (val) => {
				return val.tag === option.tag
			})
			return search === -1
		},
		selectedCategory(catID) {
			this.fetchPost('/ajax/topic/list', {
				id: catID
			}).then(res => {
				this.topicList = res.topicList
			}).catch(() => {
				return
			})
		},
		selectedTag(option) {
			let checkWord = /^[a-zA-Zก-๙0-9]+$/.test(option.tag)
			if (!checkWord) {
				this.invalidText = true
				return
			}
			if (this.tagSelected.length >= this.config.limitTag) {
				this.noMoreAdd = true
				return
			}
			this.invalidText = false
			this.noMoreAdd = false
			if (!this.searchDupTag(option)) {
				this.onEscKeyTag()
				return
			}
			this.tagSelected.push(option)
			this.onEscKeyTag()
		},
		deleteTag(index) {
			this.noMoreAdd = false
			this.tagSelected.splice(index, 1)
		},
		onDownKeyTag() {
			if (this.currentTag === '') {
				this.currentTag = 0
				return
			}
			this.currentTag += 1
			if (this.currentTag > this.tagList.length - 1) {
				this.currentTag -= 1
				return
			}
			if (this.currentTag > 2) {
				this.$refs.sws_dropdown.scrollTop += 36
			}
		},
		onUpKeyTag() {
			if (this.currentTag === '') return
			this.currentTag -= 1
			if (this.currentTag < 0) this.currentTag = 0
			if (this.currentTag < (this.tagList.length) - 2) {
				this.$refs.sws_dropdown.scrollTop -= 36
			}
		},
		onDeleteKeyTag() {
			this.invalidText = false
			this.noMoreAdd = false
			if (this.searchTag) return
			this.tagSelected.splice(-1, 1)
		},
		onEnterKeyTag() {
			if (this.hasTagList && this.currentTag !== '') {
				const option = this.tagList[this.currentTag]
				this.selectedTag({
					id: option.id + '',
					tag: option.name
				})
				return
			}
			if (!this.searchTag) return
			this.selectedTag({
				id: '0',
				tag: this.searchTag
			}, false)
		},
		seletedGap(options) {
			if (options === '') return
			this.gapMain = options
		},

		// === tag, province ===

		checkDraft: _.debounce(function () {
			this.isLoading = true
			this.descriptionPost = editor.getData()
			this.isLoading = false
		}, 700),
		initEditor() {
			let self = this
			ClassicEditor
				.create(document.querySelector('#editor'), {
					toolbar: {
						viewportTopOffset: 50,
					},
					simpleUpload: {
						uploadUrl: '/ajax/post/image'
					},
					autosave: {
						save(editor) {
							self.descriptionPost = editor.getData()
							return Promise.resolve()
						}
					}
				})
				.then(editor => {
					window.editor = editor;
				})
				.catch(err => {
					console.error(err.stack);
				});
		},
	},
})
