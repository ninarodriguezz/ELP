decoder String =
    field "definition" String

decoder List =
    list decoder String

decoder Meaning =
    map2 Meaning
        (field "partOfSpeech" String)
        (field "definitions" list)

type alias Meaning =
    {partOfSpeech : Sring,
    definition : String}

type alias Definition =
    {word : String,
    definition : List Meaning}    