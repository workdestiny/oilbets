package repository

import (
	"math/rand"
	"strings"
	"time"

	"github.com/workdestiny/oilbets/entity"
)

//AlgorithmReplacePost sub post owner index next index
func AlgorithmReplacePost(post []*entity.DiscoverModel) []*entity.DiscoverModel {

	var successPost []*entity.DiscoverModel

	var startIndexRandom int
	random := false

	for i := 0; i < len(post); i++ {

		if i == len(post)-1 {

			if !random {
				successPost = append(successPost, post[i])
				break
			}

			successPost = append(successPost, post[randomIndexPost(startIndexRandom, i)])
			break
		}

		if post[i].Owner.ID != post[i+1].Owner.ID {

			if !random {
				successPost = append(successPost, post[i])
				continue
			}

			successPost = append(successPost, post[randomIndexPost(startIndexRandom, i)])
			random = false
			continue

		}

		tm := post[i].CreatedAt.Add(time.Hour * 7)

		values := strings.Split(tm.String(), " ")

		if len(values) > 0 {

			tmNext := post[i+1].CreatedAt.Add(time.Hour * 7)
			valuesNext := strings.Split(tmNext.String(), " ")

			if len(valuesNext) > 0 {
				if values[0] == valuesNext[0] {
					if !random {
						random = true
						startIndexRandom = i
					}
					continue
				}

				if !random {
					successPost = append(successPost, post[i])
					continue
				}

				successPost = append(successPost, post[randomIndexPost(startIndexRandom, i)])
				random = false
				continue

			}
		}
	}

	return successPost

}

func randomIndexPost(min, max int) int {
	return rand.Intn((max+1)-min) + min

}
