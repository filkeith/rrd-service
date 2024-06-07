local function map_record(record)
    return map {timestamp = record.timestamp}
end

local function reduce_min(a, b)
    if a.timestamp < b.timestamp then
        return a
    else
        return b
    end
end

function find_oldest(stream)
    return stream : map(map_record) : reduce(reduce_min)
end