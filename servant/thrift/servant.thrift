namespace go   fun.rpc
namespace py   fun.rpc
namespace php  fun.rpc
namespace java fun.rpc

exception TCacheMissed {
    11: optional string message
}

exception TMongoNotFound {
    11: optional string message
}

exception TMongoDegrade {
    11: optional string message
}

exception TMongoReadOnly {
    11: optional string message
}

exception TIdTimeBackwards {
}

struct TMemcacheData {
    1: required binary data
    2: required i32 flags
}

struct Context {
    /**
     * e,g. POST+/facebook/getPaymentRequestId/+34ca2cf6
     */
    1:required string caller

    /**
     * Where the request originated.
     */
    11:optional string host

    /**
     * Remote user IP address.
     */
    12:optional string ip

    /**
     * Session id.
     */
    13:optional string sid
}

/**
 * Thrift don't support service multiplex, so we have to bury all
 * services into the giant FunServant.
 *
 * We don't want to use different port for different service for 
 * multiplex of service, that will lead to complexity for client.
 */
service FunServant {
    /**
     * Ping.
     *
     * @return string - always 'pong'
     */
    string ping(
        1: required Context ctx
    ),

    /**
     *
     */
    i64 id_next(
        1: required Context ctx
        2: i16 flag
    ) throws (
        1: TIdTimeBackwards backwards
    ),

    /**
     * Write a dlog event.
     *
     * timestamp will be generated by servant.
     *
     * @param Context ctx - Request context
     * @param string ident - Log filename
     * @param string tag -
     * @param string json - Client is responsible to jsonize
     */
    void dlog(
        /** request context */
        1: required Context ctx, 
        2: required string ident, 
        3: required string tag, 
        4: required string json
    ),

    //=================
    // lcache section
    //=================

    bool lc_set(
        1: required Context ctx, 
        2: required string key, 
        3: required binary value
    ),

    binary lc_get(
        1: required Context ctx, 
        2: required string key
    ) throws (
        1: TCacheMissed miss
    ),

    void lc_del(
        1: required Context ctx, 
        2: required string key
    ),

    //=================
    // kvdb section
    //=================
    bool kvdb_set(
        1: required Context ctx,
        2: required binary key,
        3: required binary value
    ),

    binary kvdb_get(
        1: required Context ctx,
        2: required binary key
    ),

    bool kvdb_del(
        1: required Context ctx,
        2: required binary key
    ),

    //=================
    // memcache section
    //=================

    /**
     * Set.
     *
     * @param Context ctx - Request context info.
     * @param string key -
     * @param TMemcacheData value -
     * @param i32 expiration - in seconds: either a relative time from now (up to 1 month), or 
     *     an absolute Unix epoch time. Zero means the Item has no expiration time.
     */
    bool mc_set(
        1: required Context ctx, 
        2: required string key, 
        3: required TMemcacheData value, 
        4: required i32 expiration
    ),

    /**
     * Get.
     *
     * @param Context ctx - Request context info.
     * @param string key -
     * @return TMemcacheData - Value of the key
     */
    TMemcacheData mc_get(
        1: required Context ctx, 
        2: required string key
    ) throws (
        1: TCacheMissed miss
    ),

    /**
     * Add.
     *
     * @param Context ctx - Request context info.
     * @param string key -
     * @param TMemcacheData value - Value of the key
     * @param i32 expiration -
     * @return bool - False if the key already exists.
     */
    bool mc_add(
        1: required Context ctx, 
        2: required string key, 
        3: required TMemcacheData value, 
        4: required i32 expiration
    ),

    /**
     * Delete.
     *
     * @param Context ctx - Request context info.
     * @param string key -
     * @return bool - True if Success 
     */
    bool mc_delete(
        1: required Context ctx, 
        2: required string key
    ),

    /**
     * Increment.
     *
     * @param Context ctx - Request context info.
     * @param string key -
     * @param i32 delta - If negative, means decrement
     * @return binary - New value of the key
     */
    i32 mc_increment(
        1: required Context ctx, 
        2: required string key, 
        3: required i32 delta
    ),

    //=================
    // mongodb section
    // use binary for 
    // all bson codec
    //=================

    binary mg_find_one(
        1: required Context ctx, 
        2: string pool,
        3: string table,
        4: i32 shardId,
        /** where condition */
        5: binary query,
        6: binary fields
    ) throws (
        1: TMongoNotFound miss
    ),

    list<binary> mg_find_all(
        1: required Context ctx, 
        2: string pool,
        3: string table,
        4: i32 shardId,
        5: binary query,
        6: binary fields,
        7: i32 limit,
        8: i32 skip,
        9: list<string> orderBy
    ),

    binary mg_find_id(
        1: required Context ctx,
        2: string pool,
        3: string table,
        4: i32 shardId,
        5: binary id
    ),

    i32 mg_count(
        1: required Context ctx,
        2: string pool,
        3: string table,
        4: i32 shardId,
        5: binary query
    ),

    bool mg_update(
        1: required Context ctx, 
        2: string pool,
        3: string table,
        4: i32 shardId,
        5: binary query,
        6: binary change
    ),

    bool mg_update_id(
        1: required Context ctx, 
        2: string pool,
        3: string table,
        4: i32 shardId,
        5: i32 id,
        6: binary change
    ),

    bool mg_upsert(
        1: required Context ctx, 
        2: string pool,
        3: string table,
        4: i32 shardId,
        5: binary query,
        6: binary change
    ),

    bool mg_upsert_id(
        1: required Context ctx, 
        2: string pool,
        3: string table,
        4: i32 shardId,
        5: i32 id,
        6: binary change
    ),

    bool mg_insert(
        1: required Context ctx, 
        2: string pool,
        3: string table,
        4: i32 shardId,
        5: binary doc
    ),

    bool mg_inserts(
        1: required Context ctx, 
        2: string pool,
        3: string table,
        4: i32 shardId,
        5: list<binary> docs
    ),

    bool mg_delete(
        1: required Context ctx, 
        2: string pool,
        3: string table,
        4: i32 shardId,
        5: binary query
    ),

    binary mg_find_and_modify(
        1: required Context ctx, 
        2: string pool,
        3: string table,
        4: i32 shardId,
        5: binary query,
        6: binary change,
        7: bool upsert,
        8: bool remove,
        9: bool returnNew
    ),

}
