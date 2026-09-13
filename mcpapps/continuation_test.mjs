// Public MCP Apps transport simulation; this is not a live ChatGPT host test.
import test from 'node:test';
import assert from 'node:assert/strict';
import {readFileSync} from 'node:fs';
import vm from 'node:vm';

const html=readFileSync(new URL('./continuation.html',import.meta.url),'utf8');
const script=html.match(/<script>([\s\S]*?)<\/script>/)[1];
const identity={work_session_id:'ws_1',endpoint_id:'ep_1',controller_generation:1,binding_secret:'secret_1'};
const flush=async()=>{for(let i=0;i<30;i++)await Promise.resolve()};

function host(options={}){
  let now=100000,nextTimer=0,phase='awaiting',wakeStatus='pending',paused=false;
  const timers=new Map(),windowEvents={},documentEvents={},calls=[],elements={};
  const scope={...identity};
  const response=(extra={})=>({...scope,state:{enabled:!paused,phase,rounds_used:0,max_rounds:3},wake:{state:wakeStatus},...extra});
  const handlers={};
  const elementsIDs=['title','status','scope','budget','detail','enable','pause'];
  for(const id of elementsIDs)elements[id]={textContent:'',disabled:false,addEventListener:(event,handler)=>{handlers[id+':'+event]=handler}};
  const document={visibilityState:options.hidden?'hidden':'visible',documentElement:{style:{},getBoundingClientRect:()=>({height:240})},getElementById:id=>elements[id],addEventListener:(type,handler)=>{documentEvents[type]=handler}};
  function deliver(message){windowEvents.message({source:parent,data:{jsonrpc:'2.0',...message}})}
  function reply(request,result){deliver({id:request.id,result})}
  const parent={postMessage:request=>{
    calls.push(request);
    if(!request.id)return;
    if(options.intercept&&options.intercept(request,api))return;
    if(request.method==='ui/initialize')reply(request,{protocolVersion:'2026-01-26',hostCapabilities:options.capabilities||{serverTools:{},message:{text:{}}},hostContext:{locale:'en'}});
    else if(request.method==='ui/message')reply(request,options.messageResult||{});
    else if(request.method==='tools/call'){
      const {name,arguments:args}=request.params;
      if(name==='work_continuation_pause')paused=true;
      let result=response();
      if(name==='work_wake_acquire'&&wakeStatus==='pending'){
        wakeStatus='claimed';result=response({wake_id:'wake_1',attempt_id:'attempt_1'});
      }else if(name==='work_wake_prepare'){
        wakeStatus='prepared';result=response({wake_id:'wake_1',attempt_id:'attempt_1',automatic_message:'SERVER MESSAGE\n{"consume_token":"one-use"}'});
      }else if(name==='work_wake_finish'){
        if(wakeStatus!=='consumed')wakeStatus=args.delivery_status;
        result=response();
      }
      reply(request,{structuredContent:result});
    }
  }};
  const fakeDate=class extends Date{static now(){return now}};
  const api={calls,elements,scope,response,reply,deliver,
    setWake:value=>{wakeStatus=value},setEnabled:value=>{paused=!value},setPhase:value=>{phase=value},
    toolCalls:name=>calls.filter(call=>call.method==='tools/call'&&call.params.name===name),
    messages:()=>calls.filter(call=>call.method==='ui/message'),
    notification:(method,params)=>deliver({method,params}),
    present:(data=scope)=>deliver({method:'ui/notifications/tool-result',params:{structuredContent:{...data}}}),
    click:async id=>{handlers[id+':click']();await flush()},
    visibility:async value=>{document.visibilityState=value;documentEvents.visibilitychange();await flush()},
    close:()=>windowEvents.pagehide(),
    advance:async ms=>{
      const target=now+ms;
      for(let i=0;i<1000;i++){
        const next=[...timers].filter(([,v])=>v.when<=target).sort((a,b)=>a[1].when-b[1].when)[0];
        if(!next){now=target;await flush();return}
        now=next[1].when;timers.delete(next[0]);next[1].fn();await flush();
      }
      throw new Error('Timer loop');
    }
  };
  const context=vm.createContext({window:{parent,addEventListener:(type,handler)=>{windowEvents[type]=handler}},document,navigator:{language:'en'},crypto:{randomUUID:()=>options.bindingID||'view_1'},Uint8Array,Date:fakeDate,setTimeout:(fn,delay)=>{const id=++nextTimer;timers.set(id,{fn,when:now+delay});return id},clearTimeout:id=>timers.delete(id)});
  vm.runInContext(script,context);
  return api;
}
async function start(options={}){const h=host(options);await flush();h.present();await h.advance(0);return h}
async function enable(h){await h.click('enable');await h.advance(0)}

test('requires explicit click and sends only the server message through public ui/message',async()=>{
  const h=await start();await h.advance(6000);
  assert.equal(h.messages().length,0);assert.equal(h.toolCalls('work_wake_acquire').length,0);
  assert.equal(h.toolCalls('work_continuation_bind')[0].params.arguments.user_enabled,false);
  await enable(h);
  assert.equal(h.toolCalls('work_continuation_bind')[1].params.arguments.user_enabled,true);
  assert.equal(h.messages().length,1);
  assert.equal(h.messages()[0].params.role,'user');
  assert.equal(h.messages()[0].params.content[0].text,'SERVER MESSAGE\n{"consume_token":"one-use"}');
  assert.equal(h.toolCalls('work_wake_finish')[0].params.arguments.delivery_status,'dispatch_accepted');
  await h.advance(20000);assert.equal(h.messages().length,1);
});

test('missing message text or serverTools capability disables all tool calls',async()=>{
  for(const capabilities of [{serverTools:{}},{message:{text:{}}},{serverTools:{},message:{image:{}}}]){
    const h=await start({capabilities});await h.advance(20000);
    assert.equal(h.calls.filter(c=>c.method==='tools/call').length,0);
    assert.equal(h.elements.enable.disabled,true);assert.match(h.elements.status.textContent,/does not support/);
  }
});

test('complete tool input and result work in either order without duplicate binding',async()=>{
  for(const resultFirst of [true,false]){
    const h=host();await flush();
    const input=()=>h.notification('ui/notifications/tool-input',{arguments:identity});
    if(resultFirst){h.present();input()}else{input();h.present()}
    await h.advance(0);assert.equal(h.toolCalls('work_continuation_bind').length,1);
  }
});

test('a partial input cannot supply or mix binding credentials',async()=>{
  const h=host();await flush();
  h.notification('ui/notifications/tool-input-partial',{arguments:identity});
  h.notification('ui/notifications/tool-input',{arguments:{work_session_id:identity.work_session_id}});
  h.present({endpoint_id:identity.endpoint_id,controller_generation:1,binding_secret:identity.binding_secret});
  await h.advance(10000);assert.equal(h.toolCalls('work_continuation_bind').length,0);
});

test('lost prepare response crosses the fence once and never sends or prepares again',async()=>{
  const h=await start({intercept:(request,api)=>{
    if(request.params?.name==='work_wake_prepare'){api.setWake('prepared');return true}
  }});
  await enable(h);await h.advance(45000);
  assert.equal(h.toolCalls('work_wake_prepare').length,1);assert.equal(h.messages().length,0);
  assert.equal(h.toolCalls('work_wake_finish')[0].params.arguments.delivery_status,'delivery_unknown');
  assert.match(h.elements.status.textContent,/inspection/);
});

test('isError is rejection whereas message timeout is delivery_unknown',async()=>{
  const rejected=await start({messageResult:{isError:true}});await enable(rejected);
  assert.equal(rejected.toolCalls('work_wake_finish')[0].params.arguments.delivery_status,'delivery_rejected');
  const unknown=await start({intercept:request=>request.method==='ui/message'});await enable(unknown);await unknown.advance(45000);
  assert.equal(unknown.toolCalls('work_wake_finish')[0].params.arguments.delivery_status,'delivery_unknown');
  assert.equal(unknown.messages().length,1);
});

test('hidden views do not bind or dispatch and bound hidden views only read and heartbeat',async()=>{
  const h=await start({hidden:true});await h.advance(6000);assert.equal(h.toolCalls('work_continuation_bind').length,0);
  await h.visibility('visible');await h.advance(0);
  await h.visibility('hidden');const startIndex=h.calls.length;await h.advance(20000);
  const names=h.calls.slice(startIndex).filter(c=>c.method==='tools/call').map(c=>c.params.name);
  assert.ok(names.length>0);assert.ok(names.every(name=>['work_continuation_state','work_continuation_heartbeat'].includes(name)));
  assert.equal(h.messages().length,0);
});

test('hiding after prepare prevents ui/message and defers finish until visible',async()=>{
  let prepare;
  const h=await start({intercept:request=>{if(request.params?.name==='work_wake_prepare'){prepare=request;return true}}});
  await enable(h);assert.ok(prepare);
  await h.visibility('hidden');h.reply(prepare,{structuredContent:h.response({automatic_message:'do not send'})});await flush();
  assert.equal(h.messages().length,0);assert.equal(h.toolCalls('work_wake_finish').length,0);
  await h.visibility('visible');await h.advance(0);
  assert.equal(h.messages().length,0);assert.equal(h.toolCalls('work_wake_finish')[0].params.arguments.delivery_status,'delivery_unknown');
});

test('stale generation responses cannot dispatch after explicit presentation replacement',async()=>{
  let prepare;
  const h=await start({intercept:request=>{if(request.params?.name==='work_wake_prepare'){prepare=request;return true}}});
  await enable(h);assert.ok(prepare);
  h.scope.controller_generation=2;h.scope.binding_secret='secret_2';h.present();
  h.reply(prepare,{structuredContent:{...identity,automatic_message:'stale'}});await flush();await h.advance(2000);
  assert.equal(h.messages().length,0);
  assert.equal(h.elements.enable.disabled,false);
  const latest=h.toolCalls('work_continuation_bind').at(-1).params.arguments;
  assert.equal(latest.controller_generation,2);assert.equal(latest.binding_secret,'secret_2');
});

test('late old presentation cannot downgrade the controller generation',async()=>{
  const h=await start();h.scope.controller_generation=2;h.scope.binding_secret='secret_2';h.present();await h.advance(0);
  h.present(identity);await h.advance(4000);
  assert.equal(h.toolCalls('work_continuation_bind').at(-1).params.arguments.controller_generation,2);
});

test('consume before finish remains pending model settlement without another message',async()=>{
  const h=await start({intercept:(request,api)=>{
    if(request.method==='ui/message'){api.setWake('consumed');api.reply(request,{});return true}
  }});await enable(h);await h.advance(20000);
  assert.equal(h.messages().length,1);assert.match(h.elements.detail.textContent,/consume and settle/);
});

test('pause while prepare is in flight prevents sending and confirms server pause',async()=>{
  let prepare;
  const h=await start({intercept:request=>{if(request.params?.name==='work_wake_prepare'){prepare=request;return true}}});
  await enable(h);await h.click('pause');h.reply(prepare,{structuredContent:h.response({automatic_message:'do not send'})});await flush();await h.advance(2000);
  assert.equal(h.messages().length,0);assert.equal(h.toolCalls('work_continuation_pause').length,1);
});

test('reconcile remains single-flight and backs off on disconnected state requests',async()=>{
  let stateRequest;
  const h=await start({intercept:request=>{if(request.params?.name==='work_continuation_state'){stateRequest=request;return true}}});
  assert.ok(stateRequest);
  await h.visibility('hidden');await h.visibility('visible');await h.advance(9999);
  assert.equal(h.toolCalls('work_continuation_state').length,1);
  await h.advance(1);await h.advance(3999);assert.equal(h.toolCalls('work_continuation_state').length,1);
  await h.advance(1);assert.equal(h.toolCalls('work_continuation_state').length,2);
});

test('teardown cancels timers and pending requests without new messages',async()=>{
  const h=await start();h.close();const count=h.calls.length;await h.advance(60000);assert.equal(h.calls.length,count);
});


test('presenter metadata binds without exposing secret in model-visible output',async()=>{
  const h=host();await flush();
  h.notification('ui/notifications/tool-result',{structuredContent:{work_session_id:'ws_1'},_meta:{'agentdock/work-continuation':identity}});
  await h.advance(0);assert.equal(h.toolCalls('work_continuation_bind')[0].params.arguments.binding_secret,'secret_1');
});

test('a mismatched prepare response identity never reaches the host conversation',async()=>{
  const h=await start({intercept:(request,api)=>{
    if(request.params?.name==='work_wake_prepare'){
      api.reply(request,{structuredContent:api.response({controller_generation:9,automatic_message:'stale'})});return true;
    }
  }});await enable(h);assert.equal(h.messages().length,0);assert.match(h.elements.status.textContent,/no longer bound/);
});

test('a finish timeout retries only finish, never delivery or prepare',async()=>{
  let reports=0;
  const h=await start({intercept:request=>{
    if(request.params?.name==='work_wake_finish'&&++reports===1)return true;
  }});await enable(h);await h.advance(30000);
  assert.equal(h.toolCalls('work_wake_prepare').length,1);assert.equal(h.messages().length,1);
  assert.equal(h.toolCalls('work_wake_finish').length,2);
});


test('reopened accepted wake is not acquired or prepared even after user consent',async()=>{
  const h=await start();h.setWake('dispatch_accepted');await enable(h);await h.advance(10000);
  assert.equal(h.toolCalls('work_wake_acquire').length,0);assert.equal(h.toolCalls('work_wake_prepare').length,0);
  assert.equal(h.messages().length,0);
});

test('an older endpoint presentation cannot roll back a recovered WorkSession',async()=>{
  const h=await start();h.scope.endpoint_id='ep_2';h.scope.controller_generation=2;h.scope.binding_secret='secret_2';h.present();await h.advance(0);
  h.present(identity);await h.advance(4000);
  assert.equal(h.toolCalls('work_continuation_bind').at(-1).params.arguments.endpoint_id,'ep_2');
});


test('authoritative consume resolves unknown delivery without another message',async()=>{
  const h=await start({intercept:request=>request.method==='ui/message'});await enable(h);await h.advance(10000);
  assert.match(h.elements.status.textContent,/inspection/);
  h.setWake('consumed');h.setPhase('processing');await h.advance(2000);
  assert.match(h.elements.status.textContent,/Continuation requested/);assert.equal(h.messages().length,1);
});


test('pause clearly requires model enable before this view can consent again',async()=>{
  const h=await start();await h.click('pause');await h.advance(0);
  assert.equal(h.elements.enable.disabled,true);assert.match(h.elements.detail.textContent,/assistant to enable/);
  const binds=h.toolCalls('work_continuation_bind').length;await enable(h);
  assert.equal(h.toolCalls('work_continuation_bind').length,binds);
  h.setEnabled(true);await h.advance(2000);assert.equal(h.elements.enable.disabled,false);
  await enable(h);assert.equal(h.messages().length,1);
});

test('accepted finish clears a prepared observation after a lost finish response',async()=>{
  let reports=0;
  const h=await start({intercept:request=>{
    if(request.params?.name==='work_wake_finish'&&++reports===1)return true;
  }});await enable(h);await h.advance(30000);
  assert.equal(h.messages().length,1);assert.match(h.elements.status.textContent,/Continuation requested/);
  assert.doesNotMatch(h.elements.detail.textContent,/will not resend/);
});
