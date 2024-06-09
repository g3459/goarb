contract Arouter{

    struct Route{
        uint amOut;
        bytes calls;
    }

    struct Routes{
        uint amIn;
        Route[] routes;
    }

    function findPools(address[] calldata tokens)public returns(bytes[][] memory pools){
        unchecked {
            pools=new bytes[][](tokens.length);
            for (uint t0; t0 < tokens.length; t0++)
                pools[t0]=new bytes[](tokens.length);
            for (uint t0; t0 < tokens.length; t0++){
                address token0=tokens[t0];
                for (uint t1; t1 < tokens.length; t1++){
                    address token1=tokens[t1];
                    if(token0<token1){
                        bytes32 smp;
                        assembly{
                            smp:=mload(0x40)
                            mstore(0x40,add(smp,0x20))
                        }
                        mstoreUniV2Pool(token0,token1,0x5757371414417b8C6CAad45bAeF941aBc7d3Ab32,0x96e8ac4277198ff8b6f785478aa9a39f403cb768dd02cbee326c3e7da348845f);
                        mstoreUniV2Pool(token0,token1,0xc35DADB65012eC5796536bD9864eD8773aBc74C4,0xe18a34eb0e04b04f7a0ac29a6e80748dca96319b42c54d679cb821dca90c6303);
                        mstoreUniV3Pool(token0,token1,0x1F98431c8aD98523631AE4a59f267346ea31F984,0xe34f199b19b2b4f47f68442619d555527d244f78a3297ea89325f843f87b8b54,100,1);
                        mstoreUniV3Pool(token0,token1,0x1F98431c8aD98523631AE4a59f267346ea31F984,0xe34f199b19b2b4f47f68442619d555527d244f78a3297ea89325f843f87b8b54,500,10);
                        mstoreUniV3Pool(token0,token1,0x1F98431c8aD98523631AE4a59f267346ea31F984,0xe34f199b19b2b4f47f68442619d555527d244f78a3297ea89325f843f87b8b54,3000,60);
                        mstoreUniV3Pool(token0,token1,0x1F98431c8aD98523631AE4a59f267346ea31F984,0xe34f199b19b2b4f47f68442619d555527d244f78a3297ea89325f843f87b8b54,10000,200);
                        bytes memory _pools;
                        assembly{
                            _pools:=smp
                            mstore(_pools,sub(sub(mload(0x40),smp),0x20))
                        }
                        pools[t1][t0]=pools[t0][t1]=_pools;
                    }
                }
            }
        }
    }

    function mstoreUniV2Pool(address t0,address t1, address factory, bytes32 poolInitCode) internal{
        unchecked{
            bytes32 fmp;
            assembly{fmp:=mload(0x40)}
            address pool=address(uint160(uint(keccak256(abi.encodePacked(hex'ff',factory, keccak256(abi.encodePacked(t0, t1)) ,poolInitCode)))));
            if(pool.code.length>0){
                bytes32 stateHash;uint reserve0; uint reserve1;
                {
                    (, bytes memory state) = pool.call(abi.encodeWithSelector(IUniV2Pool.getReserves.selector));
                    stateHash=keccak256(state);
                    (reserve0, reserve1)=abi.decode(state,(uint,uint));
                }
                if(reserve0>0&&reserve1>0){
                    assembly{
                        mstore(fmp,or(shl(128,reserve0),reserve1))
                        fmp:=add(fmp,0x20)
                        mstore(fmp,or(and(stateHash,0xffffffff00000000000000000000000000000000000000000000000000000000),or(shl(216,1),or(shl(160,997000),and(pool,0x000000000000000000000000ffffffffffffffffffffffffffffffffffffffff)))))
                        fmp:=add(fmp,0x20)
                    }
                }
            }
            assembly{mstore(0x40,fmp)}
        }
    }

    function mstoreUniV3Pool(address t0,address t1, address factory, bytes32 poolInitCode,uint fee,int s)internal {
        unchecked{
            bytes32 fmp;
            assembly{fmp:=mload(0x40)}
            address pool=address(uint160(uint(keccak256(abi.encodePacked(hex'ff',factory, keccak256(abi.encode(t0, t1,fee)),poolInitCode)))));
            if(pool.code.length>0){
                uint liquidity=IUniV3Pool(pool).liquidity();
                if(liquidity>2){
                    uint reserve0;uint reserve1;uint reserve0Limit;uint reserve1Limit;bytes32 stateHash;
                    {
                        (, bytes memory state) = pool.call(abi.encodeWithSelector(IUniV3Pool.slot0.selector));
                        stateHash=keccak256(state);
                        int t;
                        uint sqrtPX64;
                        (sqrtPX64,t) = abi.decode(state, (uint, int));
                        sqrtPX64>>=32;
                        reserve0=(liquidity<<64)/sqrtPX64;
                        reserve1=(liquidity*sqrtPX64)>>64;
                        if(reserve0>0&&reserve1>0){
                            reserve0Limit=(liquidity<<64)/tSqrtPX64(t < 0 ? ((t + 1) / s - 1) * s : (t / s) * s) - reserve0;
                            reserve1Limit=((liquidity*tSqrtPX64(t < 0 ? ((t + 1) / s) * s : (t / s + 1) * s))>>64) - reserve1;
                        }
                    }
                    if(reserve0>reserve0Limit&&reserve1>reserve1Limit){
                        assembly{
                            mstore(fmp,or(shl(128,reserve0),reserve1))
                            fmp:=add(fmp,0x20)
                            mstore(fmp,or(and(stateHash,0xffffffff00000000000000000000000000000000000000000000000000000000),or(shl(160,sub(1000000,fee)),and(pool,0x000000000000000000000000ffffffffffffffffffffffffffffffffffffffff))))
                            fmp:=add(fmp,0x20)
                            mstore(fmp,or(shl(128,reserve0Limit),reserve1Limit))
                            fmp:=add(fmp,0x20)
                        }
                    }
                }
            }
            assembly{mstore(0x40,fmp)}
        }
    }

    function allTokensWithBalances(address[] calldata tokens,uint[] calldata ethPricesX64,uint minEth) public returns (Routes[][] memory routes){
        unchecked{
            routes=new Routes[][](tokens.length);
            bytes[][] memory pools=findPools(tokens);
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
                    routes[i][k].routes[i].amOut=b-(b*22000)/gasPQ;
                    findRoutes(tokens,routes[i][k].routes,pools,gasPQ);
                    b>>=1;
                }
            }
        }
    }

    function singleToken(address[] calldata tokens,uint ethPriceInX64,uint amIn,uint tIn) public returns (Route[] memory routes){
        unchecked{
            routes=new Route[](tokens.length);
            uint gasPQ=((amIn*ethPriceInX64) / (tx.gasprice>0?tx.gasprice:(block.basefee+30e9)))>>64;
            routes[tIn].amOut=amIn-(amIn*22000)/gasPQ;
            bytes[][] memory pools=findPools(tokens);
            findRoutes(tokens,routes,pools,gasPQ);
        }
    }

    function findRoutes(address[] calldata tokens,Route[] memory routes,bytes[][] memory pools,uint gasPQ) internal pure{
        unchecked{
            uint updated=type(uint).max<<tokens.length;
            while(true){
                for (uint t0; t0 < tokens.length; t0++){
                    if(updated&(1<<t0)==0){
                        updated|=(1<<t0);
                        if(routes[t0].amOut>0){
                            Route memory routeIn = routes[t0];
                            for (uint t1; t1 < tokens.length; t1++){
                                if(t0!=t1){
                                    Route memory routeOut = routes[t1];
                                    bool direc = tokens[t0] < tokens[t1];
                                    bytes memory _pools=pools[t0][t1];
                                    uint p;
                                    while(p<_pools.length){
                                        uint amIn=routeIn.amOut;
                                        uint slot0;uint slot1;uint slot2;
                                        assembly{
                                            slot0:=add(add(_pools,0x20),p)
                                            slot1:=mload(add(add(_pools,0x40),p))
                                        }
                                        if(uint8(uint(slot1)>>216)==0){
                                            assembly{slot2:=mload(add(add(_pools,0x60),p))}
                                            p+=0x60;
                                        }else{
                                            slot2=type(uint).max;
                                            p+=0x40;
                                        }
                                        if((direc?(slot2>>128):uint128(slot2))>amIn && !checkPool(routeIn.calls,address(uint160(slot1)))){
                                            assembly{
                                                slot0:=mload(slot0)
                                            }
                                            amIn-=(amIn*95000)/gasPQ;
                                            uint amOut=amIn*uint24(slot1>>160);
                                            amOut = (direc
                                                ? (amOut * uint128(slot0)) / ((slot0>>128) * 1e6 + amOut)
                                                : (amOut * (slot0>>128)) / (uint128(slot0) * 1e6 + amOut));
                                            if(amOut>routeOut.amOut){
                                                assembly{updated:=and(updated,not(shl(t1,0x01)))}
                                                routeOut.amOut=amOut;
                                                bytes memory rInCalls=routeIn.calls;
                                                bytes memory rOutCalls=routeOut.calls;
                                                uint rLen=rInCalls.length;
                                                while(rLen>0){
                                                    bytes32 _call;
                                                    assembly {_call := mload(add(rInCalls, rLen))}
                                                    if(_call!=0){
                                                        break;
                                                    }
                                                    rLen-=0x20;
                                                }
                                                rLen+=0x20;
                                                if(rLen>rOutCalls.length){
                                                    rOutCalls=(routeOut.calls=new bytes(rLen));
                                                }else{
                                                    for(uint i=rLen;i<rOutCalls.length;i+=0x20){
                                                        assembly{mstore(add(add(rOutCalls,0x20),i),0)}
                                                    }
                                                }
                                                for(uint i=0x20;i<=rInCalls.length;i+=0x20){
                                                    assembly{mstore(add(rOutCalls,i),mload(add(rInCalls,i)))}
                                                }
                                                uint rsh;
                                                while(uint48(amIn)!=amIn){
                                                    rsh+=8;
                                                    amIn>>=8;
                                                }
                                                amIn<<=8;
                                                amIn|=rsh;
                                                uint callbytes=(slot1&0x7fffffffff00000000000000ffffffffffffffffffffffffffffffffffffffff)|(amIn<<160);
                                                if(direc) callbytes|=0x8000000000000000000000000000000000000000000000000000000000000000;
                                                assembly{mstore(add(rOutCalls,rLen),callbytes)}
                                            }
                                        }
                                    }
                                }
                            }
                        }
                    }
                }
                if(updated==type(uint).max) return;
            }
        }
    }

    function tSqrtPX64(int tick) public pure returns (uint sqrtPriceX64) {
        unchecked {
            uint256 absTick;
            assembly {
                let mask := sar(255, tick)
                absTick := xor(mask, add(mask, tick))
            }
            uint256 price;
            assembly {
                price := xor(shl(128, 1), mul(xor(shl(128, 1), 0xfffcb933bd6fad37aa2d162d1a594001), and(absTick, 0x1)))
            }
            if (absTick & 0x2 != 0) price = (price * 0xfff97272373d413259a46990580e213a) >> 128;
            if (absTick & 0x4 != 0) price = (price * 0xfff2e50f5f656932ef12357cf3c7fdcc) >> 128;
            if (absTick & 0x8 != 0) price = (price * 0xffe5caca7e10e4e61c3624eaa0941cd0) >> 128;
            if (absTick & 0x10 != 0) price = (price * 0xffcb9843d60f6159c9db58835c926644) >> 128;
            if (absTick & 0x20 != 0) price = (price * 0xff973b41fa98c081472e6896dfb254c0) >> 128;
            if (absTick & 0x40 != 0) price = (price * 0xff2ea16466c96a3843ec78b326b52861) >> 128;
            if (absTick & 0x80 != 0) price = (price * 0xfe5dee046a99a2a811c461f1969c3053) >> 128;
            if (absTick & 0x100 != 0) price = (price * 0xfcbe86c7900a88aedcffc83b479aa3a4) >> 128;
            if (absTick & 0x200 != 0) price = (price * 0xf987a7253ac413176f2b074cf7815e54) >> 128;
            if (absTick & 0x400 != 0) price = (price * 0xf3392b0822b70005940c7a398e4b70f3) >> 128;
            if (absTick & 0x800 != 0) price = (price * 0xe7159475a2c29b7443b29c7fa6e889d9) >> 128;
            if (absTick & 0x1000 != 0) price = (price * 0xd097f3bdfd2022b8845ad8f792aa5825) >> 128;
            if (absTick & 0x2000 != 0) price = (price * 0xa9f746462d870fdf8a65dc1f90e061e5) >> 128;
            if (absTick & 0x4000 != 0) price = (price * 0x70d869a156d2a1b890bb3df62baf32f7) >> 128;
            if (absTick & 0x8000 != 0) price = (price * 0x31be135f97d08fd981231505542fcfa6) >> 128;
            if (absTick & 0x10000 != 0) price = (price * 0x9aa508b5b7a84e1c677de54f3e99bc9) >> 128;
            if (absTick & 0x20000 != 0) price = (price * 0x5d6af8dedb81196699c329225ee604) >> 128;
            if (absTick & 0x40000 != 0) price = (price * 0x2216e584f5fa1ea926041bedfe98) >> 128;
            if (absTick & 0x80000 != 0) price = (price * 0x48a170391f7dc42444e8fa2) >> 128;
            assembly {
                if sgt(tick, 0) { price := div(not(0), price) }
                sqrtPriceX64 := shr(64, price)
            }
        }
    }


    function checkPool(bytes memory calls, address pool) internal pure returns (bool) {
        unchecked {
            for (uint i; i < calls.length; i += 0x20) {
                address _pool;
                assembly {
                    _pool := mload(add(add(calls, 0x20), i))
                }
                if (pool == _pool) return true;
            }
            return false;
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


