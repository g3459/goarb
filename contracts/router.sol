contract Arouter{

    struct Route{
        uint amOut;
        bytes calls;
    }

    struct Routes{
        uint amIn;
        Route[] routes;
    }
    
    struct UniProtocol{
        uint24[] fees;
        address factory;
        bytes32 poolInitCode;
    }

    // function findPoolsTest()public view returns(bytes[][] memory pools){
    //     address[] memory tokens=new address[](2);
    //     tokens[0]=0xdAC17F958D2ee523a2206206994597C13D831ec7;
    //     tokens[1]=0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48;
    //     UniProtocol[] memory protocols=new UniProtocol[](1);
    //     protocols[0].fees=new uint24[](1);
    //     protocols[0].fees[0]=uint24(100);
    //     protocols[0].factory=0x1F98431c8aD98523631AE4a59f267346ea31F984;
    //     protocols[0].poolInitCode=bytes32(0xe34f199b19b2b4f47f68442619d555527d244f78a3297ea89325f843f87b8b54);
    //     return this.findPools(tokens,protocols);
    // }


    function findPools(address[] calldata tokens, UniProtocol[] calldata protocols)public view returns(bytes[][] memory pools){
        unchecked {
            pools=new bytes[][](tokens.length);
            for (uint t0; t0 < tokens.length; t0++)
                pools[t0]=new bytes[](tokens.length);
            bytes32 fmp;
            assembly{
                fmp:=mload(0x40)
            }
            for (uint t0; t0 < tokens.length; t0++){
                for (uint t1; t1 < tokens.length; t1++){
                    if(tokens[t0]<tokens[t1]){
                        bytes32 stp;
                        assembly{
                            stp:=fmp
                            fmp:=add(fmp,0x20)
                        }
                        for (uint p; p < protocols.length; p++) {
                            for (uint f; f < protocols[p].fees.length; f++) {
                                UniProtocol memory protocol=protocols[p];
                                address pool=address(uint160(uint(keccak256(abi.encodePacked(hex'ff',protocol.factory, keccak256(abi.encode(tokens[t0], tokens[t1],protocol.fees[f])),protocol.poolInitCode)))));
                                if (pool.code.length > 0) {
                                    uint liquidity=IUniV3Pool(pool).liquidity();
                                    if(liquidity>2){
                                        uint reserve0;uint reserve1;uint reserve0Limit;uint reserve1Limit;bytes4 stateHash;
                                        {
                                            (, bytes memory state) = pool.staticcall(abi.encodeWithSelector(IUniV3Pool.slot0.selector));
                                            stateHash=bytes4(keccak256(state));
                                            int t;
                                            int s = IUniV3Pool(pool).tickSpacing();
                                            {
                                                uint sqrtPX64;
                                                (sqrtPX64,t) = abi.decode(state, (uint, int));
                                                sqrtPX64>>=32;
                                                reserve0=(liquidity<<64)/sqrtPX64;
                                                reserve1=(liquidity*sqrtPX64)>>64;
                                            }
                                            reserve0Limit=((liquidity<<64)/tSqrtPX64(t < 0 ? int((t + 1) / s - 1) * s : int(t / s) * s)) - reserve0;
                                            reserve1Limit=((liquidity*tSqrtPX64(t < 0 ? int((t + 1) / s) * s : int(t / s + 1) * s))>>64) - reserve1;
                                        }
                                        if(reserve0>(reserve0Limit<<1)&&reserve1>(reserve1Limit<<1)){
                                            uint fee=1e6-protocol.fees[f];
                                            assembly{
                                                mstore(fmp,or(shl(128,reserve0),reserve1))
                                                fmp:=add(fmp,0x20)
                                                mstore(fmp,or(shl(128,reserve0Limit),reserve1Limit))
                                                fmp:=add(fmp,0x20)
                                                mstore(fmp,or(or(stateHash,shl(160,fee)),pool))
                                                fmp:=add(fmp,0x20)
                                            }
                                        }
                                    }
                                }
                            }
                        }
                        bytes memory _pools;
                        assembly{
                            _pools:=stp
                            mstore(_pools,sub(sub(fmp,stp),0x20))
                        }
                        pools[t0][t1]=_pools;

                    }
                }
            }
        }
    }

    function allTokensWithBalances(address[] calldata tokens,UniProtocol[] calldata protocols,uint[] calldata ethPricesX64,uint minEth) public view returns (Routes[][] memory routes){
        unchecked{
            routes=new Routes[][](tokens.length);
            bytes[][] memory pools=findPools(tokens,protocols);
            for(uint i;i<tokens.length;i++){
                uint b = IERC20(tokens[i]).balanceOf(msg.sender);
                uint n;
                {
                    uint _b=b;
                    uint minEthX64=uint(minEth)<<64;
                    while(_b*ethPricesX64[i]>minEthX64){
                        n++;
                        _b>>=1;
                    }
                }
                routes[i]=new Routes[](n);
                for(uint k; k<n;k++){
                    routes[i][k].amIn=b;
                    routes[i][k].routes=new Route[](tokens.length);
                    uint gasPQ=((b*ethPricesX64[i]) / (tx.gasprice>0?tx.gasprice:(block.basefee+30e9)))>>64;
                    routes[i][k].routes[i].amOut=b-(b*21000)/gasPQ;
                    findRoutes(tokens,routes[i][k].routes,/*k>0?routes[i][k-1].routes:new Route[](0),*/pools,gasPQ);
                    b>>=1;
                }
            }
        }
    }

    function singleToken(address[] calldata tokens,UniProtocol[] calldata protocols,uint ethPriceInX64,uint amIn,uint tIn) public view returns (Route[] memory routes){
        unchecked{
            routes=new Route[](tokens.length);
            uint gasPQ=((amIn*ethPriceInX64) / (tx.gasprice>0?tx.gasprice:(block.basefee+30e9)))>>64;
            routes[tIn].amOut=amIn-(amIn*21000)/gasPQ;
            bytes[][] memory pools=findPools(tokens,protocols);
            findRoutes(tokens,routes,pools,gasPQ);
        }
    }

    function findRoutes(address[] calldata tokens,Route[] memory routes,/*Route[] memory routes2,*/bytes[][] memory pools,uint gasPQ) internal pure{
        unchecked{
            bytes32 updated;
            while(true){
                for (uint t0; t0 < tokens.length; t0++){
                    updated|=bytes32(0x01<<t0);
                    if(routes[t0].amOut>0){
                        for (uint t1; t1 < tokens.length; t1++){
                            if(t0!=t1){
                                Route memory routeIn = routes[t0];
                                Route memory routeOut = routes[t1];
                                bool direc = tokens[t0] < tokens[t1];
                                    bytes memory _pools=pools[t0][t1];
                                    for(uint p;p<_pools.length;p+=32){
                                    uint amOut=routeIn.amOut;
                                    uint128 reserve0;uint128 reserve1;address pool;
                                    assembly{
                                        reserve1:=mload(add(add(_pools,32),p))
                                        reserve0:=shr(128,reserve1)
                                        pool:=mload(add(add(_pools,96),p))
                                    }
                                    if((direc?reserve0:reserve1)>amOut<<3 && !checkPool(routeIn.calls,pool)){
                                        uint24 fee;
                                        assembly{
                                            fee:=shr(pool,160)
                                        }
                                        amOut=(amOut-(amOut*85000)/gasPQ)*fee;
                                        amOut = (direc
                                            ? (amOut * reserve1) / (reserve0 * 1e6 + amOut)
                                            : (amOut * reserve0) / (reserve1 * 1e6 + amOut));
                                        uint128 reserve0Limit;uint128 reserve1Limit;
                                        assembly{
                                            reserve1Limit:=mload(add(add(_pools,64),p))
                                            reserve0Limit:=shr(128,reserve1Limit)
                                        }
                                        if((direc?reserve1Limit:reserve0Limit)>amOut){
                                            if(amOut>routeOut.amOut){
                                                routeOut.amOut=amOut;
                                                bytes memory rInCalls=routeIn.calls;
                                                bytes memory rOutCalls=routeOut.calls;
                                                uint rLen;
                                                while(rLen<rInCalls.length){
                                                    bytes8 _call;
                                                    assembly {
                                                        _call := mload(add(add(rInCalls, 0x20), rLen))
                                                    }
                                                    if(_call==0){
                                                        break;
                                                    }
                                                    rLen+=24;
                                                }
                                                if(rLen+24>rOutCalls.length){
                                                    rOutCalls=(routeOut.calls=new bytes((((rLen-1)>>5)<<5)+32));
                                                }else{
                                                    for(uint i=rLen+32;i<rOutCalls.length+64;i+=32){
                                                        assembly{
                                                            mstore(add(rOutCalls,i),0)
                                                        }
                                                    }
                                                }
                                                
                                                for(uint i=32;i<rLen+64;i+=32){
                                                    assembly{
                                                        mstore(add(rOutCalls,i),mload(add(rInCalls,i)))
                                                    }
                                                }
                                                bytes32 callbytes=bytes20(pool);
                                                bytes4 stateHash;
                                                assembly{
                                                    stateHash:=pool
                                                }
                                                direc?stateHash|=bytes4(0x00000001):stateHash&=bytes4(0xfffffffe);
                                                callbytes|=bytes24(stateHash)>>160;
                                                assembly{
                                                    mstore(add(add(rOutCalls,0x20),rLen),callbytes)
                                                }
                                            }
                                        }
                                    }
                                }
                            }
                        }
                    }
                }
                if(updated==bytes32(uint(0x01<<tokens.length)-1)) return;
            }
        }
    }

    function checkPool(bytes memory calls, address pool) internal pure returns (bool) {
        unchecked {
            for (uint i; i < calls.length; i += 24) {
                address _pool;
                assembly {
                    _pool := mload(add(add(calls, 0x20), i))
                }
                if (pool == _pool) return true;
            }
            return false;
        }
    }

    // function test()public view returns(bytes24 a,bytes24 b,bytes memory c){
    //     a=bytes20(address(this));
    //     b=a|bytes24(bytes4(0xffffffff))>>160;
    //     // assembly{
    //     //     a:=mload(0x40)
    //     //     b:=add(a,0x20)
    //     //     mstore(b,b)
    //     //     b:=add(b,0x20)
    //     //     mstore(b,b)
    //     //     b:=add(b,0x20)
    //     //     mstore(b,b)
    //     //     b:=add(b,0x20)
    //     //     mstore(b,b)
    //     //     b:=add(b,0x20)
    //     // }
    //     // bytes memory bb;
    //     // assembly{
    //     //     mstore(a,sub(b,add(a,0x20)))
    //     //     bb:=a
    //     // }
    //     // c=bb;
        
    
    // }

    function tSqrtPX64(int t) internal pure returns(uint) {
        unchecked{
            uint abst = t < 0 ? uint(-t) : uint(t);
            uint ratio = abst & 0x1 != 0 ? 0xfffcb933bd6fad37aa2d162d1a594001 : 0x100000000000000000000000000000000;
            if (abst & 0x2 != 0) ratio = (ratio * 0xfff97272373d413259a46990580e213a) >> 128;
            if (abst & 0x4 != 0) ratio = (ratio * 0xfff2e50f5f656932ef12357cf3c7fdcc) >> 128;
            if (abst & 0x8 != 0) ratio = (ratio * 0xffe5caca7e10e4e61c3624eaa0941cd0) >> 128;
            if (abst & 0x10 != 0) ratio = (ratio * 0xffcb9843d60f6159c9db58835c926644) >> 128;
            if (abst & 0x20 != 0) ratio = (ratio * 0xff973b41fa98c081472e6896dfb254c0) >> 128;
            if (abst & 0x40 != 0) ratio = (ratio * 0xff2ea16466c96a3843ec78b326b52861) >> 128;
            if (abst & 0x80 != 0) ratio = (ratio * 0xfe5dee046a99a2a811c461f1969c3053) >> 128;
            if (abst & 0x100 != 0) ratio = (ratio * 0xfcbe86c7900a88aedcffc83b479aa3a4) >> 128;
            if (abst & 0x200 != 0) ratio = (ratio * 0xf987a7253ac413176f2b074cf7815e54) >> 128;
            if (abst & 0x400 != 0) ratio = (ratio * 0xf3392b0822b70005940c7a398e4b70f3) >> 128;
            if (abst & 0x800 != 0) ratio = (ratio * 0xe7159475a2c29b7443b29c7fa6e889d9) >> 128;
            if (abst & 0x1000 != 0) ratio = (ratio * 0xd097f3bdfd2022b8845ad8f792aa5825) >> 128;
            if (abst & 0x2000 != 0) ratio = (ratio * 0xa9f746462d870fdf8a65dc1f90e061e5) >> 128;
            if (abst & 0x4000 != 0) ratio = (ratio * 0x70d869a156d2a1b890bb3df62baf32f7) >> 128;
            if (abst & 0x8000 != 0) ratio = (ratio * 0x31be135f97d08fd981231505542fcfa6) >> 128;
            if (abst & 0x10000 != 0) ratio = (ratio * 0x9aa508b5b7a84e1c677de54f3e99bc9) >> 128;
            if (abst & 0x20000 != 0) ratio = (ratio * 0x5d6af8dedb81196699c329225ee604) >> 128;
            if (abst & 0x40000 != 0) ratio = (ratio * 0x2216e584f5fa1ea926041bedfe98) >> 128;
            if (abst & 0x80000 != 0) ratio = (ratio * 0x48a170391f7dc42444e8fa2) >> 128;
            if (t > 0) ratio = type(uint).max / ratio;
            return ratio >> 64;
        }
    }
}



//interfaces

interface IUniV2Pool{
    function swap(uint amount0Out, uint amount1Out, address to, bytes calldata data) external;
    function getReserves()external view returns(uint reserve0, uint reserve1, uint blockTimestampLast);
}

interface IUniV3Pool{
    function swap(address recipient, bool zeroForOne, int amountSpecified, uint160 sqrtPriceLimitX96, bytes calldata data) external returns(int amount0, int amount1);
    function fee() external view returns(uint24 fee);
    function slot0() external view returns(uint sqrtPX96, int t, uint observationIndex, uint observationCardinality, uint observationCardinalityNext, uint feeProtocol, bool unlocked);
    function liquidity() external view returns(uint liquidity);
    function tickSpacing() external view returns(int s);
    function flash(address recipient, uint amount0, uint amount1, bytes calldata data) external;
}

interface IAlgebraV3Pool{
    function swap(address recipient, bool zeroForOne, int amountSpecified, uint160 sqrtPriceLimitX96, bytes calldata data) external returns(int amount0, int amount1);
    function globalState() external view returns(uint sqrtPX96, int t, uint fee, uint timePointIndex, uint comunityFeet0, uint comunityFeeT1, bool unlocked);
    function tickSpacing() external view returns(int s);
    function liquidity() external view returns(uint liquidity);
}

interface IERC20{
    function balanceOf(address ) external view returns ( uint256 );
    function transfer(address, uint256 ) external returns ( uint256 );
}


