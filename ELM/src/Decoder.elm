module Decoder exposing (..)

import Http
import Json.Decode exposing (..)

type alias Meaning =
    {partOfSpeech : String,
    definition : (List String)}

type alias Definition =
    { word : String
    , definition : List Meaning}    


decoderJson : Decoder (List Definition)
decoderJson =
  list decoderDef

decoderDef : Decoder Definition
decoderDef =
  map2 Definition
    (field "word" string)
    (field "meanings" (list decoderMeaning))

decoderMeaning : Decoder Meaning
decoderMeaning =
  map2 Meaning
    (field "partOfSpeech" string)
    (field "definitions" (list decoderString))

decoderString : Decoder String 
decoderString =
    field "definition" string

