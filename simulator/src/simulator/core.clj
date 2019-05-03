(ns simulator.core
  (:require [clj-docker-client.core :as docker])
  (:gen-class))

(comment

  ;; test cases
  ;; start-registry
  ;; download-packages
  ;; download-packages
  ;; second download packages is faster than first

  ;; start-registry
  ;; start-registry
  ;; download-packages from first
  ;; download-packages from second
  ;; second download packages is faster than first

  ;; start five registries with share off
  ;; download packages
  ;; start five registries with share on
  ;; download packages
  ;; second version is faster

  (create-nodes 1 {:registry true})

  (defscenario cached-faster {:nodes 1 })

)

(def scenarios (atom []))

(defmacro defscenario [title & forms])

;; open-registry is the central index
;; bolivar is a local federation proxy that runs a registry that can also share

;; start open-registry (fresh index) (should have control over index?)
;; start bolivar
;; run test-project that downloads directly from open-registry
;; run test-project that downloads via bolivar, keep running
;; second test-project that downloads via bolivar

;; todo
;; - [ ] image for open-registry
;; - [ ] rename open-registry-fed -> bolivar
;; - [ ] methods for
;;  - [ ] running open-registry
;;  - [ ] running bolivar
;;  - [ ] running commands in existing container

(def test-net [:open-registry 1
               :bolivar 2
               :deps-test 2])

(def meas1 (measure (install-deps (test-net :deps-test 1))))
(def meas2 (measure (install-deps (test-net :deps-test 2))))
(is (< meas1 meas1))

;; dsl for testing and actions in that system

;; model system for properties

;; create container with node:12.0.0-stretch
;; run open-registry binary in it
;; git clone webpack
;; rm yarn.lock
;; set registry
;; run yarn install and measure executing time

(comment
  (def conn (docker/connect))

  (with-open [conn (docker/connect)]
    (docker/ping conn))

  ;; (docker/pull conn "node:12.0.0-stretch")

  (def container-id (docker/create conn "open-registry-fed"))

  (docker/start conn container-id)

  (println container-id)

  (docker/logs conn container-id)

  (docker/exec conn container-id "whoami")

  (def container-ip  (-> (docker/inspect conn container-id)
                              :NetworkSettings
                              :Networks
                              :bridge
                              :IPAddress
                              ))
  "172.17.0.2"

  )

(defn -main
  "I don't do a whole lot ... yet."
  [& args]
  (println "Hello, World!"))
