import axios from 'axios'

export function fetchPost(url, data) {
  return fetch(url, {
      credentials: 'include',
      method: 'POST',
      body: JSON.stringify(data)
    }).then(function (res) {
      if (res.ok && res.status == 200) {
        return res.json()
      }
      throw Error(res.statusText)
    })
}

export function fetchPostStatus(url, data) {
  return axios.post(url, data, {
    withCredentials: true
  })
}